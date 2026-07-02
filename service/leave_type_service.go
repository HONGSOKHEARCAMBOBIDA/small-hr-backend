package service

import (
	"context"
	"errors"
	"mysql/config"
	"mysql/helper"
	"mysql/model"
	"mysql/request"
	"mysql/response"

	"gorm.io/gorm"
)

type LeaveTypeService interface {
	GetLeaveTypes(ctx context.Context, userID int) ([]response.LeaveTypeResponse, error)
	CreateLeaveType(ctx context.Context, input request.LeaveTypeRequestCreate) error
	UpdateLeaveType(ctx context.Context, id int, input request.LeaveTypeRequestUpdate) error
}

type leavetypeservice struct {
	db *gorm.DB
}

func NewLeaveTypeService() LeaveTypeService {
	return &leavetypeservice{
		db: config.DB,
	}
}

func (s *leavetypeservice) GetLeaveTypes(ctx context.Context, userID int) ([]response.LeaveTypeResponse, error) {
	var data []response.LeaveTypeResponse
	var user model.User
	if err := s.db.WithContext(ctx).First(&user, userID).Error; err != nil {
		return nil, err
	}

	query := s.db.WithContext(ctx).Table("leave_type lt").
		Select(`
			lt.id AS id,
			c.id AS company_id,
			c.name AS company_name,
			lt.code AS code,
			lt.name AS name,
			lt.is_active AS is_active,
			lt.is_deduct AS is_deduct
		`).
		Joins("LEFT JOIN company c ON c.id = lt.company_id")

	query = helper.ManageCompanyFilter(query, s.db, user)

	if err := query.Scan(&data).Error; err != nil {
		return nil, err
	}
	return data, nil
}

func (s *leavetypeservice) CreateLeaveType(ctx context.Context, input request.LeaveTypeRequestCreate) error {
	tx := s.db.Begin()
	if tx.Error != nil {
		return tx.Error
	}

	newleavetype := model.LeaveType{
		CompanyID: input.CompanyID,
		Code:      input.Code,
		Name:      input.Name,
		Isactive:  true,
		IsDeduct:  input.IsDeduct,
	}

	if err := tx.WithContext(ctx).Create(&newleavetype).Error; err != nil {
		tx.Rollback()
		return err
	}
	return tx.Commit().Error
}

func (s *leavetypeservice) UpdateLeaveType(ctx context.Context, id int, input request.LeaveTypeRequestUpdate) error {
	updates := map[string]interface{}{}

	if input.CompanyID != nil {
		updates["company_id"] = *input.CompanyID
	}
	if input.Code != nil {
		updates["code"] = *input.Code
	}
	if input.Name != nil {
		updates["name"] = *input.Name
	}
	if input.Isactive != nil {
		updates["is_active"] = *input.Isactive
	}
	if input.IsDeduct != nil {
		updates["is_deduct"] = *input.IsDeduct
	}

	if len(updates) == 0 {
		return errors.New("no field to update")
	}

	result := s.db.WithContext(ctx).Model(&model.LeaveType{}).Where("id =?", id).Updates(updates)
	if result.Error != nil {
		return result.Error
	}
	return nil
}
