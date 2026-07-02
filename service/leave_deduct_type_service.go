package service

import (
	"context"
	"mysql/config"
	"mysql/model"

	"gorm.io/gorm"
)

type LeaveDeductTypeService interface {
	GetLeaveDeductType(ctx context.Context) ([]model.LeaveDeductType, error)
}

type leavededucttypeservice struct {
	db *gorm.DB
}

func NewLeaveDeductTypeService() LeaveDeductTypeService {
	return &leavededucttypeservice{
		db: config.DB,
	}
}

func (s *leavededucttypeservice) GetLeaveDeductType(ctx context.Context) ([]model.LeaveDeductType, error) {
	var data []model.LeaveDeductType

	if err := s.db.WithContext(ctx).Find(&data).Error; err != nil {
		return nil, err
	}

	return data, nil
}
