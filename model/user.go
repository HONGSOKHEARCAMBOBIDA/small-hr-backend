package model

import "mysql/model/base"

type User struct {
	base.ModelBase
	PhoneHash    string `json:"phone_hash" gorm:"column:phone_hash"`
	PasswordHash string `json:"password_hash" gorm:"column:password_hash"`
	RoleID       int    `json:"role_id" gorm:"column:role_id"`
	Role         Role
	IsActive     bool   `json:"is_active" gorm:"column:is_active"`
	Name         string `json:"name" gorm:"column:name"`
	Gender       int    `json:"gender" gorm:"column:gender"`
	BaseSalary   string `json:"base_salary" gorm:"column:base_salary"`
	CompanyID    int    `json:"company_id" gorm:"column:company_id"`
	Company      Company
	QrToken      string `json:"qr_token" gorm:"column:qr_token"`
	IsVerify     bool   `json:"is_verify" gorm:"column:is_verify"`
}

func (User) TableName() string {
	return "user"
}
