package model

import "mysql/model/base"

type UserCompany struct {
	base.ModelBase
	UserID    int `json:"user_id" gorm:"column:user_id"`
	CompanyID int `json:"company_id" gorm:"column:company_id"`
}

func (UserCompany) TableName() string {
	return "user_company"
}
