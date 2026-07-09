package service

import (
	"context"
	"errors"
	"fmt"
	"log"
	"mysql/config"
	"mysql/helper"
	"mysql/model"
	"mysql/request"
	"mysql/response"
	"mysql/utils"
	"strings"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type LeaveRequestService interface {
	CreateLeaveRequest(ctx context.Context, id int, input request.LeaveRequestCreate) error
	UpdateLeaveRequest(ctx context.Context, id int, userID int, input request.LeaveRequestUpdate) error
	UpdateStatusLeaveRequest(ctx context.Context, user_id int, id int, input request.LeaveRequestUpdateStatus) error
	GetLeaveRequest(ctx context.Context, id int, pf request.Pagination, filter map[string]string) ([]response.LeaveRequestResponse, *model.PaginationMetadata, error)
	DeleteLeaveRequest(ctx context.Context, id int) error
}

type leaveRequestService struct {
	db *gorm.DB
}

func NewLeaveRequestService() LeaveRequestService {
	return &leaveRequestService{
		db: config.DB,
	}
}

const (
	GenderMale   = 1
	GenderFemale = 2

	LeaveStatusPending = 1

	telegramSendTimeout = 10 * time.Second
	dataLayout          = "2006-01-02"

	ManageOneCompany      = 1
	ManageMultipleCompany = 2
	ManageAllCompany      = 3

	RoleLevelStaft     = 1
	RoleLevelManager   = 2
	RoleLevelDeveloper = 7

	defaultPageSize = 10
	maxPageSize     = 20
)

func validateLeaveRequestInput(input request.LeaveRequestCreate) error {
	start, err := time.Parse(dataLayout, input.StartDate)
	if err != nil {
		return fmt.Errorf("កាលបរិច្ឆេទចាប់ផ្តើមមិនត្រឹមត្រូវ៖ %w", err)
	}
	end, err := time.Parse(dataLayout, input.EndDate)
	if err != nil {
		return fmt.Errorf("កាលបរិច្ឆេទបញ្ចប់មិនត្រឹមត្រូវ %w", err)
	}
	backToWork, err := time.Parse(dataLayout, input.BackToWorkDate)
	if err != nil {
		return fmt.Errorf("កាលបរិច្ឆេទត្រឡប់ទៅធ្វើការវិញមិនត្រឹមត្រូវ %w", err)
	}
	if end.Before(start) {
		return errors.New("កាលបរិច្ឆេទបញ្ចប់មិនត្រូវមុនកាលបរិច្ឆេទចាប់ផ្តើមទេ")
	}
	if backToWork.Before(end) {
		return errors.New("កាលបរិច្ឆេទត្រឡប់ទៅធ្វើការវិញមិនត្រូវមុនកាលបរិច្ឆេទបញ្ចប់ទេ")
	}
	if input.TotalDay <= 0 {
		return errors.New("ចំនួនថ្ងៃសរុបត្រូវតែធំជាងសូន្យ")
	}
	if input.Reason != nil {
		trimmed := strings.TrimSpace(*input.Reason)
		if trimmed == "" {
			return errors.New("អ្នកត្រូវបញ្ចូលមូលហេតុ")
		}
		if len(trimmed) < 3 {
			return errors.New("ហេតុផលខ្លីពេក")
		}
		if len(trimmed) > 500 {
			return errors.New("ហេតុផលវេងពេក")
		}
	}
	return nil
}

func hasOverlappingLeaveRequest(tx *gorm.DB, userID int, startDate, endDate string) (bool, error) {
	var count int64
	err := tx.Model(&model.LeaveRequest{}).
		Where("user_id =?", userID).
		Where("status = ?", LeaveStatusPending).
		Where("start_date <= ? AND end_date >= ?", endDate, startDate).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func wrapNotFound(err error, entity string, id int) error {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return fmt.Errorf("%s %d not found: %w", entity, id, err)
	}
	return fmt.Errorf("failed to load %s:%w", entity, err)
}

func genderLabel(gender int) string {
	switch gender {
	case GenderMale:
		return "ខ្ញុំបាទ"
	case GenderFemale:
		return "នាងខ្ញុំ"
	default:
		return "ខ្ញុំ"
	}
}

func approverGenerLable(gender int) string {
	switch gender {
	case GenderMale:
		return "លោកគ្រូ"
	case GenderFemale:
		return "អ្នកគ្រូ"
	default:
		return "លោកគ្រូ/អ្នកគ្រូ"
	}
}

func buildLeaveRequestMessage(user, approve model.User, deduction model.LeaveDeductType, input request.LeaveRequestCreate) string {
	requester := genderLabel(user.Gender)
	approveGender := approverGenerLable(approve.Gender)

	return fmt.Sprintf(
		"<b>សូមជម្រាបសួរលោកគ្រូ អ្នកគ្រូ!</b>\n\n"+
			"<i>%s</i> <b>%s</b> សុំអនុញ្ញាតច្បាប់ឈប់សម្រាក<b>%v</b>%s ចាប់ពីថ្ងៃទី%s ដល់ថ្ងៃទី%s ចូលបម្រើការងារវិញនៅថ្ងៃទី%s។\n"+
			"<code>*មូលហេតុ :%s។\n</code>"+
			"សូមអធ្យាស្រ័យ%s %sជួយអនុម័តច្បាប់របស់%s ផងសូមអរគុណ 🙏",
		requester,
		user.Name,
		input.TotalDay,
		deduction.Name,
		input.StartDate,
		input.EndDate,
		input.BackToWorkDate,
		*input.Reason,
		approveGender,
		approve.Name,
		requester,
	)
}

func notifyApprover(message, groupChatID, botToken string) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("recovered panic while sending telegram message: %v", r)
		}
	}()

	ctx, cancel := context.WithTimeout(context.Background(), telegramSendTimeout)
	defer cancel()

	done := make(chan error, 1)
	go func() {
		done <- helper.SendTelegramMessage(message, groupChatID, botToken)
	}()

	select {
	case err := <-done:
		if err != nil {
			log.Printf("failed to send telegram leave notification: %v", err)
		}
	case <-ctx.Done():
		log.Printf("timed out sending telegram leave notification after %s", telegramSendTimeout)
	}
}

func (s *leaveRequestService) CreateLeaveRequest(ctx context.Context, id int, input request.LeaveRequestCreate) error {
	if err := validateLeaveRequestInput(input); err != nil {
		return fmt.Errorf("invalid leave request: %w", err)
	}

	tx := s.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return fmt.Errorf("faile to start transaction: %w", tx.Error)
	}

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()

		}
	}()

	var user model.User
	if err := tx.Preload("Company").First(&user, id).Error; err != nil {
		tx.Rollback()
		return wrapNotFound(err, "user", id)
	}

	var approver model.User
	if err := tx.First(&approver, input.ApproveBy).Error; err != nil {
		tx.Rollback()
		return wrapNotFound(err, "approver", input.ApproveBy)
	}

	var leaveType model.LeaveType
	if err := tx.First(&leaveType, input.LeaveTypeID).Error; err != nil {
		tx.Rollback()
		return wrapNotFound(err, "leaveType", input.LeaveTypeID)
	}

	var deduction model.LeaveDeductType
	if err := tx.First(&deduction, input.DeductTypeID).Error; err != nil {
		tx.Rollback()
		return wrapNotFound(err, "LeaveDeductType", input.DeductTypeID)
	}

	if user.Company.GroupChatID == nil || user.Company.BotToken == nil {
		tx.Rollback()
		return errors.New("company telegram configuration is missing")
	}

	groupChatID, err := utils.DecryptChatID(*user.Company.GroupChatID)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to decrypt groupoo chat id: %w", err)
	}

	boToken, err := utils.DecryptBotToken(*user.Company.BotToken)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to decrypt bot token: %w", err)
	}

	overlaps, err := hasOverlappingLeaveRequest(tx, id, input.StartDate, input.EndDate)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to check overlapping leave request: %w", err)
	}

	if overlaps {
		tx.Rollback()
		return errors.New("មានច្បាប់របស់អ្នកមិនទាន់អនុម័តនៅឡេីយទេ")
	}

	newLeaveRequest := model.LeaveRequest{
		UserID:         id,
		LeaveTypeID:    input.LeaveTypeID,
		StartDate:      input.StartDate,
		EndDate:        input.EndDate,
		BackToWorkDate: input.BackToWorkDate,
		TotalDay:       input.TotalDay,
		DeductTypeID:   input.DeductTypeID,
		Reason:         *input.Reason,
		Status:         LeaveStatusPending,
		ApproveBy:      input.ApproveBy,
		PayrollID:      nil,
		ApproveAt:      nil,
	}

	if err := tx.Create(&newLeaveRequest).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("faile to created leave request: %w", err)
	}

	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("failed to commit leave request: %w", err)
	}

	message := buildLeaveRequestMessage(user, approver, deduction, input)
	go notifyApprover(message, groupChatID, boToken)

	return nil

}

func validateLeaveRequestUpdateInput(input request.LeaveRequestUpdate, userID int) error {
	var start, end, backToWork time.Time
	var err error

	if input.StartDate != nil {
		start, err = time.Parse(dataLayout, *input.StartDate)
		if err != nil {
			return fmt.Errorf("កាលបរិច្ឆេទចាប់ផ្តើមមិនត្រឹមត្រូវ៖ %w", err)
		}
	}
	if input.EndDate != nil {
		end, err = time.Parse(dataLayout, *input.EndDate)
		if err != nil {
			return fmt.Errorf("កាលបរិច្ឆេទបញ្ចប់មិនត្រឹមត្រូវ៖ %w", err)
		}
	}
	if input.BackToWorkDate != nil {
		backToWork, err = time.Parse(dataLayout, *input.BackToWorkDate)
		if err != nil {
			return fmt.Errorf("កាលបរិច្ឆេទត្រឡប់ទៅធ្វើការវិញមិនត្រឹមត្រូវ៖ %w", err)
		}
	}

	if input.StartDate != nil && input.EndDate != nil && end.Before(start) {
		return errors.New("កាលបរិច្ឆេទបញ្ចប់មិនត្រូវមុនកាលបរិច្ឆេទចាប់ផ្តើមទេ")
	}
	if input.EndDate != nil && input.BackToWorkDate != nil && backToWork.Before(end) {
		return errors.New("កាលបរិច្ឆេទត្រឡប់ទៅធ្វើការវិញមិនត្រូវមុនកាលបរិច្ឆេទបញ្ចប់ទេ")
	}

	if input.TotalDay != nil && *input.TotalDay <= 0 {
		return errors.New("ចំនួនថ្ងៃសរុបត្រូវតែធំជាងសូន្យ")
	}

	if input.Reason != nil {
		trimmed := strings.TrimSpace(*input.Reason)
		switch {
		case trimmed == "":
			return errors.New("អ្នកត្រូវបញ្ចូលមូលហេតុ")
		case len(trimmed) < 3:
			return errors.New("ហេតុផលខ្លីពេក")
		case len(trimmed) > 500:
			return errors.New("ហេតុផលវេងពេក")
		}
	}

	// if *input.ApproveBy == userID {
	// 	return errors.New("អ្នកខ្លួនឯងមិនអាចអនុម័តច្បាប់ខ្លួនឯងទេ")
	// }

	return nil
}

func (s *leaveRequestService) UpdateLeaveRequest(ctx context.Context, id int, userID int, input request.LeaveRequestUpdate) error {
	if id <= 0 {
		return fmt.Errorf("invalid id: %d", id)
	}

	if err := validateLeaveRequestUpdateInput(input, userID); err != nil {
		return fmt.Errorf("invalid leave request: %w", err)
	}

	updates := map[string]interface{}{}

	if input.LeaveTypeID != nil {
		updates["leave_type_id"] = *input.LeaveTypeID
	}
	if input.StartDate != nil {
		updates["start_date"] = *input.StartDate
	}
	if input.EndDate != nil {
		updates["end_date"] = *input.EndDate
	}
	if input.BackToWorkDate != nil {
		updates["back_to_work_date"] = *input.BackToWorkDate
	}
	if input.TotalDay != nil {
		updates["total_day"] = *input.TotalDay
	}
	if input.DeductTypeID != nil {
		updates["deduct_type_id"] = *input.DeductTypeID
	}
	if input.Reason != nil {
		updates["reason"] = *input.Reason
	}

	if len(updates) == 0 {
		return nil
	}

	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		result := tx.Model(&model.LeaveRequest{}).
			Where("id = ?", id).
			Updates(updates)

		if result.Error != nil {
			return fmt.Errorf("failed to update leave request %d: %w", id, result.Error)
		}
		if result.RowsAffected == 0 {
			return errors.New("no field to updated")
		}
		return nil
	})

	if err != nil {
		fmt.Errorf("failed to update leave request", "id", id, "error", err)
		return err
	}

	return nil
}

func (s *leaveRequestService) UpdateStatusLeaveRequest(ctx context.Context, user_id int, id int, input request.LeaveRequestUpdateStatus) error {
	if input.Status == nil {
		return errors.New("status is required")
	}

	tx := s.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return fmt.Errorf("faile to start transaction: %w", tx.Error)
	}

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()

		}
	}()

	var leaveRequest model.LeaveRequest

	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&leaveRequest, id).Error; err != nil {
		return err
	}

	// if leaveRequest.UserID == user_id {
	// 	return errors.New("អ្នកមិនអាចអនុម័តច្បាប់របស់ខ្លួនឯងបានទេ")
	// }

	updates := map[string]interface{}{
		"status":      *input.Status,
		"approved_at": time.Now(),
	}

	result := tx.Model(&model.LeaveRequest{}).Where("id = ?", id).Updates(updates)
	if result.Error != nil {
		return fmt.Errorf("faild to update: %w", result.Error)
	}
	// if result.RowsAffected == 0 {
	// 	return errors.New("មិនអាចធ្វើបច្ចុប្បន្នភាពស្ថានភាពបានទេ")
	// }

	return tx.Commit().Error

}

func applyAccessFilterLeaveRequest(query *gorm.DB, db *gorm.DB, role model.Role, user model.User) *gorm.DB {
	if role.Level > RoleLevelStaft && role.Level <= RoleLevelManager {
		switch user.ManageCompany {
		case ManageOneCompany:
			return query.Where("u.company_id =?", user.CompanyID)
		case ManageMultipleCompany:
			var companyIDs []int
			db.Model(&model.UserCompany{}).Where("user_id =?", user.ID).Pluck("company_id", &companyIDs)
			if len(companyIDs) == 0 {
				return query.Where("1 = 0")
			}
			return query.Where("u.company_id IN ?", companyIDs)
		case ManageAllCompany:
			return query
		default:
			return query.Where("1 = 0")
		}
	} else if role.Level <= RoleLevelStaft {
		return query.Where("u.id =?", user.ID)
	} else if role.Level > RoleLevelManager {
		return query
	}

	return query
}

func applyCommonFilterLeaveRequest(query *gorm.DB, filter map[string]string) *gorm.DB {
	for key, value := range filter {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		switch key {
		case "name":
			query = query.Where("u.name LIKE ?", "%"+helper.EscapeLike(value)+"%")
		case "company_id":
			query = query.Where("u.company_id =?", value)
		case "role_id":
			query = query.Where("u.role_id =?", value)
		case "status":
			query = query.Where("l.status =?", value)
		}
	}
	return query
}

func (s *leaveRequestService) GetLeaveRequest(ctx context.Context, id int, pf request.Pagination, filter map[string]string) ([]response.LeaveRequestResponse, *model.PaginationMetadata, error) {
	var data []response.LeaveRequestResponse
	var user model.User
	if err := s.db.WithContext(ctx).Preload("Role").First(&user, id).Error; err != nil {
		return nil, nil, err
	}

	if pf.Page < 1 {
		pf.Page = 1
	}
	if pf.PageSize < 1 {
		pf.PageSize = defaultPageSize
	} else if pf.PageSize > maxPageSize {
		pf.PageSize = maxPageSize
	}
	offset := (pf.Page - 1) * pf.PageSize

	query := s.db.WithContext(ctx).Table("leave_request l").
		Select(`
		l.id AS id,
		u.id AS user_id,
		u.gender AS gender,
		u.name AS user_name,
		r.display_name AS role_name,
		c.name AS company_name,
		lt.id AS leave_type_id,
		lt.code AS leave_type_code,
		lt.name AS leave_type_name,
		lt.is_deduct AS leave_type_is_deduct,
		l.start_date AS start_date,
		l.end_date AS end_date,
		l.back_to_work_date AS back_to_work_date,
		l.total_day AS total_day,
		ld.id AS deduct_type_id,
		ld.code AS deduct_type_code,
		ld.name AS deduct_type_name,
		l.reason AS reason,
		l.status AS status,
		ua.id AS approve_by,
		ua.name AS approve_by_name,
		l.approved_at AS approved_at
	`).
		Joins("LEFT JOIN user u ON u.id = l.user_id").
		Joins("LEFT JOIN role r ON r.id = u.role_id").
		Joins("LEFT JOIN company c ON c.id = u.company_id").
		Joins("LEFT JOIN leave_type lt ON lt.id = l.leave_type_id").
		Joins("LEFT JOIN leave_deduct_type ld ON ld.id = l.deduct_type_id").
		Joins("LEFT JOIN user ua ON ua.id = l.approve_by")

	query = applyAccessFilterLeaveRequest(query, s.db, user.Role, user)
	query = applyCommonFilterLeaveRequest(query, filter)

	var totalCount int64
	countQuery := query.Session(&gorm.Session{})
	if err := countQuery.Count(&totalCount).Error; err != nil {
		return nil, nil, err
	}

	query = query.Order("l.id DESC")
	if err := query.Offset(offset).Limit(pf.PageSize).Scan(&data).Error; err != nil {
		return nil, nil, err
	}

	for i := range data {
		data[i].StartDate = helper.FormatDate(data[i].StartDate)
		data[i].EndDate = helper.FormatDate(data[i].EndDate)
		data[i].BackToWorkDate = helper.FormatDate(data[i].BackToWorkDate)
		data[i].StatusString = helper.LeaveRequestStatus(data[i].Status)
		data[i].ApproveAt = helper.FormatDate(data[i].ApproveAt)
	}

	return data, helper.BuildPaginationMeta(pf, totalCount), nil
}

func (s *leaveRequestService) DeleteLeaveRequest(ctx context.Context, id int) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		result := tx.
			Where("id = ? AND payroll_id IS NULL", id).
			Delete(&model.LeaveRequest{})

		if result.Error != nil {
			return fmt.Errorf("failed to delete leave request: %w", result.Error)
		}

		if result.RowsAffected == 0 {
			return fmt.Errorf("leave request not found or has already been processed in payroll")
		}

		return nil
	})
}
