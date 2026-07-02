package model

import "mysql/model/base"

type LeaveType struct {
	base.ModelBase
	CompanyID int    `json:"company_id" gorm:"column:company_id"`
	Code      string `json:"code" gorm:"column:code"`
	Name      string `json:"name" gorm:"column:name"`
	Isactive  bool   `json:"is_active" gorm:"column:is_active"`
	IsDeduct  bool   `json:"is_deduct" gorm:"column:is_deduct"`
}

func (LeaveType) TableName() string {
	return "leave_type"
}
