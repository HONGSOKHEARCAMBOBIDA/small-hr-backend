package service

import (
	"context"
	"fmt"
	"mysql/config"
	"mysql/helper"
	"mysql/model"
	"mysql/request"
	"mysql/response"
	"strconv"

	"gorm.io/gorm"
)

type PayrollService interface {
	GetDraftPayroll(ctx context.Context, payrolltype int, company_id int) ([]response.PayrollDraftResponse, error)
	CreatePayroll(ctx context.Context, req request.PayrollRequestCreate) error
	GetPayroll(ctx context.Context, id int, pf request.Pagination, filter map[string]string) ([]response.PayrollResponse, *model.PaginationMetadata, error)
	DeletePayroll(ctx context.Context, req request.PayrollRequestDelete) error
}

type payrollservice struct {
	db *gorm.DB
}

func NewPayrollService() PayrollService {
	return &payrollservice{
		db: config.DB,
	}
}

func (s *payrollservice) GetDraftPayroll(ctx context.Context, payrolltype int, company_id int) ([]response.PayrollDraftResponse, error) {

	type rawRow struct {
		UserID              int
		UserName            string
		RoleName            string
		BasicSalary         string
		Currency            string
		LatePenalty         string
		LeftEarlyPenalty    string
		CompanyTotalWorkDay int
		TotalWorkDay        int
		TotalLate           int
		TotalLeftEarly      int
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
			c.total_work_day            AS company_total_work_day,
			COUNT(DISTINCT CASE WHEN             a.is_paid = false THEN a.id END) AS total_work_day,
			COUNT(DISTINCT CASE WHEN ar.attendance_type = 3 AND ar.is_permission = false AND a.is_paid = false THEN ar.id END) AS total_late,
			COUNT(DISTINCT CASE WHEN ar.attendance_type = 4 AND a.is_paid = false THEN ar.id END) AS total_left_early
		FROM user u
		LEFT JOIN role r ON r.id = u.role_id
		LEFT JOIN company c ON c.id = u.company_id
		LEFT JOIN attendance a ON a.user_id = u.id
		LEFT JOIN attendance_record ar ON ar.attendance_id = a.id
		WHERE u.company_id = ?
		  AND u.is_active   = true
		GROUP BY
			u.id, u.name, r.display_name, u.base_salary,
			c.currency, c.late_penalty, c.left_early_penalty, c.total_work_day
	`, company_id).Scan(&rows).Error

	if err != nil {
		return nil, fmt.Errorf("failed to fetch payroll draft rows: %w", err)
	}

	for i := range rows {
		decrypted, err := helper.DecryptSalary(rows[i].BasicSalary)
		if err != nil {
			return nil, err
		}
		rows[i].BasicSalary = decrypted
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

	type leaveRow struct {
		ID          int
		UserID      int
		TotalDay    float64
		DeductValue float64
	}

	var leaveRows []leaveRow
	if len(userIDs) > 0 {
		err = s.db.WithContext(ctx).Raw(`
			SELECT
				lr.id            AS id,
				lr.user_id       AS user_id,
				lr.total_day     AS total_day,
				ldt.deduct_value AS deduct_value
			FROM leave_request lr
			JOIN leave_type lt        ON lt.id = lr.leave_type_id AND lt.is_deduct = true
			JOIN leave_deduct_type ldt ON ldt.id = lr.deduct_type_id
			WHERE lr.user_id IN (?)
			  AND lr.payroll_id IS NULL
			  AND lr.status = 2
		`, userIDs).Scan(&leaveRows).Error

		if err != nil {
			return nil, fmt.Errorf("failed to fetch leave deduction records: %w", err)
		}
	}

	leaveDaysByUser := make(map[int]float64)
	unpaidleavebyUser := make(map[int][]response.CountUnPaidLeave)
	for _, l := range leaveRows {
		leaveDaysByUser[l.UserID] += l.TotalDay * l.DeductValue
		unpaidleavebyUser[l.UserID] = append(unpaidleavebyUser[l.UserID], response.CountUnPaidLeave{
			LeaveID: l.ID,
		})
	}

	payrolls := make([]response.PayrollDraftResponse, 0, len(rows))

	for _, row := range rows {
		basicSalaryF := helper.ParseFloat(row.BasicSalary)
		halfSalary := basicSalaryF / 2

		latePenaltyF := helper.ParseFloat(row.LatePenalty)
		leftEarlyPenaltyF := helper.ParseFloat(row.LeftEarlyPenalty)

		attendanceDeduction := (latePenaltyF * float64(row.TotalLate)) +
			(leftEarlyPenaltyF * float64(row.TotalLeftEarly))

		var leaveDeduction float64
		if row.CompanyTotalWorkDay > 0 {
			perDayRate := basicSalaryF / float64(row.CompanyTotalWorkDay)
			leaveDeduction = perDayRate * leaveDaysByUser[row.UserID]
		}

		totalDeduction := attendanceDeduction + leaveDeduction

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
			TotalPenaltyLate:      strconv.FormatFloat(latePenaltyF*float64(row.TotalLate), 'f', 2, 64),
			TotalLeftEarly:        row.TotalLeftEarly,
			TotalLeftEarlyPenalty: strconv.FormatFloat(leftEarlyPenaltyF*float64(row.TotalLeftEarly), 'f', 2, 64),
			TotalLeaveDay:         leaveDaysByUser[row.UserID],        // new field
			LeaveDeduction:        helper.FormatFloat(leaveDeduction), // new field
			TotalDeduction:        helper.FormatFloat(totalDeduction),
			NetSalary:             helper.FormatFloat(netSalary),
			Currency:              row.Currency,
			CountUnPaidAttendance: unpaidByUser[row.UserID],
			CountUnPaidLeave:      unpaidleavebyUser[row.UserID],
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
				if err := tx.Model(&model.Attendance{}).
					Where("id IN (?) AND user_id = ?", attendanceIDs, item.UserID).
					Updates(map[string]interface{}{
						"is_paid":    true,
						"payroll_id": payroll.ID,
					}).Error; err != nil {
					return fmt.Errorf("failed to update attendance for user %d: %w", item.UserID, err)
				}
			}
			if len(item.CountUnPaidLeave) > 0 {
				leaveIDs := make([]int, 0, len(item.CountUnPaidLeave))
				for _, r := range item.CountUnPaidLeave {
					leaveIDs = append(leaveIDs, r.LeaveID)
				}
				if err := tx.Model(&model.LeaveRequest{}).
					Where("id IN (?) AND user_id = ?", leaveIDs, item.UserID).
					Updates(map[string]interface{}{
						"payroll_id": payroll.ID,
					}).Error; err != nil {
					return fmt.Errorf("failed to update leave")
				}
			}
		}
		return nil
	})
}

func applyAccessFilterPayroll(query *gorm.DB, db *gorm.DB, role model.Role, user model.User) *gorm.DB {
	if role.Level > 1 && role.Level < 7 {
		return query.Where("u.company_id = ?", user.CompanyID)
	} else if role.Level <= 1 {
		return query.Where("u.id =?", user.ID)
	}

	return query
}

func applyCommonFilterPayroll(query *gorm.DB, filter map[string]string) *gorm.DB {
	for key, value := range filter {
		if value == "" {
			continue
		}
		switch key {
		case "name":
			query = query.Where("u.name LIKE ?", "%"+value+"%")
		case "payroll_date":
			query = query.Where("DATE_FORMAT(p.payroll_date, '%Y-%m') = ?", value)
		case "payroll_type":
			query = query.Where("p.payroll_type =?", value)
		case "company_id":
			query = query.Where("u.company_id =?", value)
		}
	}
	return query
}

func (s *payrollservice) GetPayroll(
	ctx context.Context,
	id int,
	pf request.Pagination,
	filter map[string]string,
) ([]response.PayrollResponse, *model.PaginationMetadata, error) {

	var payroll []response.PayrollResponse
	var user model.User

	if err := s.db.WithContext(ctx).
		Preload("Role").
		First(&user, id).Error; err != nil {
		return nil, nil, err
	}

	offset := (pf.Page - 1) * pf.PageSize

	baseQuery := s.db.WithContext(ctx).
		Table("payroll AS p").
		Joins("LEFT JOIN user u ON u.id = p.user_id").
		Joins("LEFT JOIN role r ON r.id = u.role_id").
		Joins("LEFT JOIN company c ON c.id = u.company_id")

	baseQuery = applyAccessFilterPayroll(baseQuery, s.db, user.Role, user)
	baseQuery = applyCommonFilterPayroll(baseQuery, filter)

	var totalCount int64
	if err := baseQuery.Session(&gorm.Session{}).
		Count(&totalCount).Error; err != nil {
		return nil, nil, err
	}

	dataQuery := baseQuery.Select(`
		p.id AS id,
		u.name AS user_name,
		u.gender AS user_gender,
		r.display_name AS role_name,
		c.name AS company_name,
		p.basic_salary AS basic_salary,
		p.half_salary AS half_salary,
		p.other AS other,
		p.total_work_day AS total_work_day,
		p.total_deduction AS total_deduction,
		p.net_salary AS net_salary,
		c.currency AS currency,
		p.payroll_type AS payroll_type,
		p.payroll_date AS payroll_date,
		p.status AS status,
		p.note AS note
	`)

	dataQuery = dataQuery.Order("p.id DESC")

	if err := dataQuery.
		Offset(offset).
		Limit(pf.PageSize).
		Scan(&payroll).Error; err != nil {
		return nil, nil, err
	}

	for i := range payroll {
		payroll[i].PayrollDate = helper.FormatDate(payroll[i].PayrollDate)
	}

	return payroll, helper.BuildPaginationMeta(pf, totalCount), nil
}

func (s *payrollservice) DeletePayroll(ctx context.Context, req request.PayrollRequestDelete) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if len(req.PayrollIDs) > 0 {
			if err := tx.Model(&model.Payroll{}).
				Where("id IN (?)", req.PayrollIDs).
				Delete(&model.Payroll{}).Error; err != nil {
				return fmt.Errorf("failed to delete payroll: %w", err)
			}

			if err := tx.Model(&model.Attendance{}).
				Where("payroll_id IN (?)", req.PayrollIDs).
				Updates(map[string]interface{}{
					"is_paid":    false,
					"payroll_id": nil,
				}).Error; err != nil {
				return fmt.Errorf("failed to update attendance: %w", err)
			}
			if err := tx.Model(&model.LeaveRequest{}).
				Where("payroll_id IN (?)", req.PayrollIDs).
				Updates(map[string]interface{}{
					"payroll_id": nil,
				}).Error; err != nil {
				return fmt.Errorf("faild to update leave")
			}
		}
		return nil
	})
}
