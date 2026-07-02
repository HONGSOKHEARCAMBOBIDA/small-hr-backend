package model

import "mysql/model/base"

type LeaveDeductType struct {
	base.ModelBase
	Code        string  `json:"code" gorm:"column:code"`
	Name        string  `json:"name" gorm:"column:name"`
	DeductValue float64 `json:"deduct_value" gorm:"column:deduct_value"`
	Isactive    bool    `json:"is_active" gorm:"column:is_active"`
}

func (LeaveDeductType) TableName() string {
	return "leave_deduct_type"
}
