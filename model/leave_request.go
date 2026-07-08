package model

import "mysql/model/base"

type LeaveRequest struct {
	base.ModelBase
	UserID          int             `json:"user_id" gorm:"column:user_id"`
	LeaveTypeID     int             `json:"leave_type_id" gorm:"column:leave_type_id"`
	StartDate       string          `json:"start_date" gorm:"column:start_date"`
	EndDate         string          `json:"end_date" gorm:"column:end_date"`
	BackToWorkDate  string          `json:"back_to_work_date" gorm:"column:back_to_work_date"`
	TotalDay        float64         `json:"total_day" gorm:"column:total_day"`
	DeductTypeID    int             `json:"deduct_type_id" gorm:"column:deduct_type_id"`
	Reason          string          `json:"reason" gorm:"column:reason"`
	Status          int             `json:"status" gorm:"column:status"`
	ApproveBy       int             `json:"approve_by" gorm:"column:approve_by"`
	PayrollID       *int            `json:"payroll_id" gorm:"column:payroll_id"`
	ApproveAt       *string         `json:"approved_at" gorm:"column:approved_at"`
	LeaveDeductType LeaveDeductType `gorm:"foreignKey:deduct_type_id"`
}

func (LeaveRequest) TableName() string {
	return "leave_request"
}
