package request

type AuthRequest struct {
	Phone    string `json:"phone"`
	Password string `json:"password"`
}

type RegisterRequest struct {
	PhoneHash  string `json:"phone_hash" gorm:"column:phone_hash"`
	RoleID     int    `json:"role_id" gorm:"column:role_id"`
	Name       string `json:"name" gorm:"column:name"`
	Gender     int    `json:"gender" gorm:"column:gender"`
	BaseSalary string `json:"base_salary" gorm:"column:base_salary"`
	CompanyID  int    `json:"company_id" gorm:"column:company_id"`
}
