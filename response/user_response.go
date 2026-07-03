package response

import (
	"mysql/model/base"
)

type UserResponse struct {
	base.ModelBase
	PhoneHash string `json:"phone_hash" gorm:"column:phone_hash"`
	//	PasswordHash string `json:"password_hash" gorm:"column:password_hash"`
	RoleID        int             `json:"role_id" gorm:"column:role_id"`
	RoleName      string          `json:"role_name" gorm:"column:role_name"`
	IsActive      bool            `json:"is_active" gorm:"column:is_active"`
	Name          string          `json:"name" gorm:"column:name"`
	Gender        int             `json:"gender" gorm:"column:gender"`
	GenderString  string          `json:"gender_string"`
	BaseSalary    string          `json:"base_salary" gorm:"column:base_salary"`
	Currency      string          `json:"currency"`
	CompanyID     int             `json:"company_id" gorm:"column:company_id"`
	CompanyName   string          `json:"company_name" gorm:"column:company_name"`
	QrToken       string          `json:"qr_token" gorm:"column:qr_token"`
	IsVerify      bool            `json:"is_verify" gorm:"column:is_verify"`
	ShiftResponse []ShiftResponse `json:"shift_response" gorm:"-"`
	ManageCompany int             `json:"manage_company"`
	CompanyIDs    []int           `json:"company_ids" gorm:"-"`
}

type UserCount struct {
	Total int `json:"total"`
}

type UserApprove struct {
	ID       int    `json:"id"`
	UserName string `json:"user_name"`
}
