package service

import (
	"context"
	"fmt"
	"mysql/config"
	"mysql/model"
	"mysql/response"
	"strconv"

	"gorm.io/gorm"
)

type PayrollService interface {
	GetDraftPayroll(ctx context.Context, payrolltype int, id int) ([]response.PayrollDraftResponse, error)
}

type payrollservice struct {
	db *gorm.DB
}

func NewPayrollService() PayrollService {
	return &payrollservice{
		db: config.DB,
	}
}

func (s *payrollservice) GetDraftPayroll(ctx context.Context, payrolltype int, id int) ([]response.PayrollDraftResponse, error) {
	// Verify the requesting user (HR) exists
	var user model.User
	if err := s.db.WithContext(ctx).First(&user, id).Error; err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	// -------------------------------------------------------------------
	// Step 1: Fetch base payroll draft data per user
	// -------------------------------------------------------------------
	//
	// attendance_record.type values (payrollroletype):
	//   1 = arrived early        (ចូលធ្វើការមុនម៉ោង)
	//   2 = arrived on time      (ចូលធ្វើការទាន់ម៉ោង)
	//   3 = arrived late         (ចូលធ្វើការយឺត)        ← late
	//   4 = left early           (ចេញពីធ្វើការមុនម៉ោង)  ← left early
	//   5 = left on time         (ចេញពីធ្វើការត្រឹមម៉ោង)
	//   6 = left after hours     (ចេញពីធ្វើការក្រោយម៉ោង)
	//
	// We count only records where attendance.is_paid = false (unpaid).
	// -------------------------------------------------------------------

	type rawRow struct {
		UserID           int
		UserName         string
		RoleName         string
		BasicSalary      string
		Currency         string
		LatePenalty      string
		LeftEarlyPenalty string
		TotalWorkDay     int
		TotalLate        int
		TotalLeftEarly   int
	}

	var rows []rawRow

	err := s.db.WithContext(ctx).Raw(`
		SELECT
			u.id                        AS user_id,
			u.name                      AS user_name,
			r.display_name              AS role_name,
			u.base_salary               AS basic_salary,
			c.currency                  AS currency,
			c.late_penalty              AS late_penalty,
			c.left_early_penalty        AS left_early_penalty,

			-- total work days: attendance rows that are not day-off
			COUNT(DISTINCT CASE WHEN a.status != 'absent' THEN a.id END) AS total_work_day,

			-- total late: unpaid attendance_record entries with type = 3
			COUNT(DISTINCT CASE WHEN ar.type = 3 AND a.is_paid = false THEN ar.id END) AS total_late,

			-- total left early: unpaid attendance_record entries with type = 4
			COUNT(DISTINCT CASE WHEN ar.type = 4 AND a.is_paid = false THEN ar.id END) AS total_left_early

		FROM user u
		LEFT JOIN role          r  ON r.id        = u.role_id
		LEFT JOIN company       c  ON c.id        = u.company_id
		LEFT JOIN attendance    a  ON a.user_id   = u.id
		LEFT JOIN attendance_record ar ON ar.attendance_id = a.id

		WHERE u.company_id = (SELECT company_id FROM user WHERE id = ?)
		  AND u.is_active   = true

		GROUP BY
			u.id, u.name, r.display_name, u.base_salary,
			c.currency, c.late_penalty, c.left_early_penalty
	`, id).Scan(&rows).Error

	if err != nil {
		return nil, fmt.Errorf("failed to fetch payroll draft rows: %w", err)
	}

	// -------------------------------------------------------------------
	// Step 2: For each user fetch their unpaid attendance IDs
	// -------------------------------------------------------------------

	// Collect all user IDs in one go
	userIDs := make([]int, 0, len(rows))
	for _, r := range rows {
		userIDs = append(userIDs, r.UserID)
	}

	type unpaidRow struct {
		UserID       int
		AttendanceID int
		Type         int // 3=late, 4=left early
	}

	var unpaidRows []unpaidRow
	if len(userIDs) > 0 {
		err = s.db.WithContext(ctx).Raw(`
			SELECT
				a.user_id        AS user_id,
				a.id             AS attendance_id,
				ar.type          AS type
			FROM attendance a
			JOIN attendance_record ar ON ar.attendance_id = a.id
			WHERE a.user_id IN (?)
			  AND a.is_paid   = false
			  AND ar.type     IN (3, 4)
		`, userIDs).Scan(&unpaidRows).Error

		if err != nil {
			return nil, fmt.Errorf("failed to fetch unpaid attendance records: %w", err)
		}
	}

	// Index unpaid rows by user_id for quick lookup
	unpaidByUser := make(map[int][]response.CountUnPaidAttendance)
	for _, u := range unpaidRows {
		unpaidByUser[u.UserID] = append(unpaidByUser[u.UserID], response.CountUnPaidAttendance{
			AttendanceID: u.AttendanceID,
			Type:         u.Type,
		})
	}

	// -------------------------------------------------------------------
	// Step 3: Assemble the final response, computing deductions
	// -------------------------------------------------------------------

	payrolls := make([]response.PayrollDraftResponse, 0, len(rows))

	for _, row := range rows {
		basicSalaryF := parseFloat(row.BasicSalary)
		halfSalary := basicSalaryF / 2

		// Calculate deductions when penalties are configured
		latePenaltyF := parseFloat(row.LatePenalty)
		leftEarlyPenaltyF := parseFloat(row.LeftEarlyPenalty)

		totalDeduction := (latePenaltyF * float64(row.TotalLate)) +
			(leftEarlyPenaltyF * float64(row.TotalLeftEarly))

		netSalary := basicSalaryF - totalDeduction
		if payrolltype == 2 { // half-month payroll
			netSalary = halfSalary - totalDeduction
		}

		payrolls = append(payrolls, response.PayrollDraftResponse{
			UserID:                row.UserID,
			UserName:              row.UserName,
			RoleName:              row.RoleName,
			BasicSalary:           row.BasicSalary,
			HalfSalary:            formatFloat(halfSalary),
			TotalWorkDay:          row.TotalWorkDay,
			TotalLate:             row.TotalLate,
			TotalLeftEarly:        row.TotalLeftEarly,
			TotalDeduction:        formatFloat(totalDeduction),
			NetSalary:             formatFloat(netSalary),
			Currency:              row.Currency,
			CountUnPaidAttendance: unpaidByUser[row.UserID],
		})
	}

	return payrolls, nil
}

// -------------------------------------------------------------------
// Helpers
// -------------------------------------------------------------------

func parseFloat(s string) float64 {
	if s == "" {
		return 0
	}
	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0
	}
	return v
}

func formatFloat(f float64) string {
	return strconv.FormatFloat(f, 'f', 2, 64)
}
