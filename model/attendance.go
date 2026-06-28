package model

import "mysql/model/base"

type Attendance struct {
	base.ModelBase
	UserID    int    `json:"user_id" gorm:"column:user_id"`
	CheckDate string `json:"check_date" gorm:"column:check_date"`
	Status    string `json:"status" gorm:"column:status"`
	IsPaid    bool   `json:"is_paid" gorm:"column:is_paid"`
	PayrollID int    `json:"payroll_id" gorm:"column:payroll_id"`
}

func (Attendance) TableName() string {
	return "attendance"
}
