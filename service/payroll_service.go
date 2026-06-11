package service

import (
	"context"
	"fmt"
	"mysql/config"
	"mysql/helper"
	"mysql/model"
	"mysql/request"
	"mysql/response"

	"gorm.io/gorm"
)

type PayrollService interface {
	GetDraftPayroll(ctx context.Context, payrolltype int, id int) ([]response.PayrollDraftResponse, error)
	CreatePayroll(ctx context.Context, req request.PayrollRequestCreate) error
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

	var user model.User
	if err := s.db.WithContext(ctx).First(&user, id).Error; err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

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
			COUNT(DISTINCT CASE WHEN a.status = 'COMPLETE' AND a.is_paid = false THEN a.id END) AS total_work_day,
			COUNT(DISTINCT CASE WHEN ar.attendance_type = 3 AND a.is_paid = false THEN ar.id END) AS total_late,
			COUNT(DISTINCT CASE WHEN ar.attendance_type = 4 AND a.is_paid = false THEN ar.id END) AS total_left_early
		FROM user u
		LEFT JOIN role r ON r.id = u.role_id
		LEFT JOIN company c ON c.id = u.company_id
		LEFT JOIN attendance a ON a.user_id = u.id
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

	userIDs := make([]int, 0, len(rows))
	for _, r := range rows {
		userIDs = append(userIDs, r.UserID)
	}

	type unpaidRow struct {
		UserID       int
		AttendanceID int
	}

	var unpaidRows []unpaidRow
	if len(userIDs) > 0 {
		err = s.db.WithContext(ctx).Raw(`
			SELECT
				a.user_id AS user_id,
				a.id AS attendance_id
			FROM attendance a
			WHERE a.user_id IN (?)
			  AND a.is_paid   = false
		`, userIDs).Scan(&unpaidRows).Error

		if err != nil {
			return nil, fmt.Errorf("failed to fetch unpaid attendance records: %w", err)
		}
	}

	unpaidByUser := make(map[int][]response.CountUnPaidAttendance)
	for _, u := range unpaidRows {
		unpaidByUser[u.UserID] = append(unpaidByUser[u.UserID], response.CountUnPaidAttendance{
			AttendanceID: u.AttendanceID,
		})
	}

	payrolls := make([]response.PayrollDraftResponse, 0, len(rows))

	for _, row := range rows {
		basicSalaryF := helper.ParseFloat(row.BasicSalary)
		halfSalary := basicSalaryF / 2

		latePenaltyF := helper.ParseFloat(row.LatePenalty)
		leftEarlyPenaltyF := helper.ParseFloat(row.LeftEarlyPenalty)

		totalDeduction := (latePenaltyF * float64(row.TotalLate)) +
			(leftEarlyPenaltyF * float64(row.TotalLeftEarly))

		netSalary := basicSalaryF - totalDeduction
		if payrolltype == 2 {
			netSalary = halfSalary - totalDeduction
		}

		payrolls = append(payrolls, response.PayrollDraftResponse{
			UserID:                row.UserID,
			UserName:              row.UserName,
			RoleName:              row.RoleName,
			BasicSalary:           row.BasicSalary,
			HalfSalary:            helper.FormatFloat(halfSalary),
			TotalWorkDay:          row.TotalWorkDay,
			TotalLate:             row.TotalLate,
			TotalLeftEarly:        row.TotalLeftEarly,
			TotalDeduction:        helper.FormatFloat(totalDeduction),
			NetSalary:             helper.FormatFloat(netSalary),
			Currency:              row.Currency,
			CountUnPaidAttendance: unpaidByUser[row.UserID],
		})
	}

	return payrolls, nil
}

func (s *payrollservice) CreatePayroll(ctx context.Context, req request.PayrollRequestCreate) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, item := range req.Payrolls {
			payroll := model.Payroll{
				UserID:         item.UserID,
				BasicSalary:    item.BasicSalary,
				HalfSalary:     item.HalfSalary,
				Other:          item.Other,
				TotalWorkDay:   item.TotalWorkDay,
				TotalDeduction: item.TotalDeduction,
				NetSalary:      item.NetSalary,
				PayrollType:    item.PayrollType,
				PayrollDate:    item.PayrollDate,
				Status:         0,
				Note:           item.Note,
			}
			if err := tx.Create(&payroll).Error; err != nil {
				return fmt.Errorf("failed to create payroll for user %d: %w", item.UserID, err)
			}
			if len(item.UnPaidAttendance) > 0 {
				attendanceIDs := make([]int, 0, len(item.UnPaidAttendance))
				for _, a := range item.UnPaidAttendance {
					attendanceIDs = append(attendanceIDs, a.AttendanceID)
				}
				if err := tx.Model(&model.Attendance{}).Where("id IN (?) AND user_id =?", attendanceIDs, item.UserID).Update("is_paid", true).Error; err != nil {
					return fmt.Errorf("failed to update attendance for user %d: %w", item.UserID, err)
				}
			}
		}
		return nil
	})
}
