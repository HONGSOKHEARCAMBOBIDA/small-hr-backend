package service

import (
	"context"
	"errors"
	"mysql/config"
	"mysql/helper"
	"mysql/model"
	"mysql/request"
	"mysql/response"
	"time"

	"gorm.io/gorm"
)

type LeaveRequestService interface {
	CreateLeaveRequest(ctx context.Context, id int, input request.LeaveRequestCreate) error
	UpdateLeaveRequest(ctx context.Context, id int, input request.LeaveRequestUpdate) error
	UpdateStatusLeaveRequest(ctx context.Context, user_id int, id int, input request.LeaveRequestUpdateStatus) error
	GetLeaveRequest(ctx context.Context, id int, pf request.Pagination, filter map[string]string) ([]response.LeaveRequestResponse, *model.PaginationMetadata, error)
}

type leaveRequestService struct {
	db *gorm.DB
}

func NewLeaveRequestService() LeaveRequestService {
	return &leaveRequestService{
		db: config.DB,
	}
}

func (s *leaveRequestService) CreateLeaveRequest(ctx context.Context, id int, input request.LeaveRequestCreate) error {
	newLeaveRequest := model.LeaveRequest{
		UserID:         id,
		LeaveTypeID:    input.LeaveTypeID,
		StartDate:      input.StartDate,
		EndDate:        input.EndDate,
		BackToWorkDate: input.BackToWorkDate,
		TotalDay:       input.TotalDay,
		DeductTypeID:   input.DeductTypeID,
		Reason:         input.Reason,
		Status:         1,
		ApproveBy:      input.ApproveBy,
		PayrollID:      nil,
		ApproveAt:      nil,
	}

	return s.db.WithContext(ctx).Create(&newLeaveRequest).Error
}

func (s *leaveRequestService) UpdateLeaveRequest(ctx context.Context, id int, input request.LeaveRequestUpdate) error {
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
	if input.ApproveBy != nil {
		updates["approve_by"] = *input.ApproveBy
	}

	if len(updates) == 0 {
		return nil
	}

	result := s.db.WithContext(ctx).
		Model(&model.LeaveRequest{}).
		Where("id = ?", id).
		Updates(updates)

	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (s *leaveRequestService) UpdateStatusLeaveRequest(ctx context.Context, user_id int, id int, input request.LeaveRequestUpdateStatus) error {
	var leaveRequest model.LeaveRequest
	if err := s.db.First(&leaveRequest, id).Error; err != nil {
		return err
	}

	if leaveRequest.UserID == user_id {
		return errors.New("អ្នកមិនអាចអនុម័តច្បាប់របស់ខ្លួនឯងបានទេ") // can't approve your own leave request
	}

	updates := map[string]interface{}{}
	if input.Status != nil {
		updates["status"] = *input.Status
		updates["approved_at"] = time.Now().Format("2006-01-02 15:04:05")
	}
	if len(updates) == 0 {
		return nil
	}

	result := s.db.WithContext(ctx).
		Model(&model.LeaveRequest{}).
		Where("id = ?", id).
		Updates(updates)

	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("មិនអាចធ្វើបច្ចុប្បន្នភាពស្ថានភាពបានទេ") // could not update status
	}
	return nil
}

func applyAccessFilterLeaveRequest(query *gorm.DB, db *gorm.DB, role model.Role, user model.User) *gorm.DB {
	if role.Level > 1 && role.Level < 7 {
		switch user.ManageCompany {
		case 1:
			return query.Where("u.company_id =?", user.CompanyID)
		case 2:
			var companyIDs []int
			db.Model(&model.UserCompany{}).Where("user_id =?", user.ID).Pluck("company_id", &companyIDs)
			if len(companyIDs) == 0 {
				return query.Where("1 = 0")
			}
			return query.Where("u.company_id IN ?", companyIDs)
		}
		return query
	} else if role.Level <= 1 {
		return query.Where("u.id =?", user.ID)
	}

	return query
}

func applyCommonFilterLeaveRequest(query *gorm.DB, filter map[string]string) *gorm.DB {
	for key, value := range filter {
		if value == "" {
			continue
		}
		switch key {
		case "name":
			query = query.Where("u.name LIKE ?", "%"+value+"%")
		case "company_id":
			query = query.Where("u.company_id = ?", value)
		case "role_id":
			query = query.Where("u.role_id = ?", value)
		case "status":
			query = query.Where("l.status = ?", value)
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
		pf.PageSize = 10
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
