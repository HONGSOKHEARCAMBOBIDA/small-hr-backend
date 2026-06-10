package model

import "mysql/model/base"

type Payroll struct {
	base.ModelBase
	UserID         int    `json:"user_id" gorm:"column:user_id"`
	BasicSalary    string `json:"basic_salary" gorm:"column:basic_salary"`
	HalfSalary     string `json:"half_salary" gorm:"column:half_salary"`
	Other          string `json:"other" gorm:"column:other"`
	TotalWorkDay   int    `json:"total_work_day" gorm:"column:total_work_day"`
	TotalDeduction string `json:"total_deduction" gorm:"column:total_deduction"`
	NetSalary      string `json:"net_salary" gorm:"column:net_salary"`
	PayrollType    int    `json:"payroll_type" gorm:"column:payroll_type"`
	PayrollDate    string `json:"payroll_date" gorm:"column:payroll_date"`
	Status         int    `json:"status" gorm:"column:status"`
	Note           string `json:"note" gorm:"column:note"`
}

func (Payroll) TableName() string {
	return "payroll"
}
