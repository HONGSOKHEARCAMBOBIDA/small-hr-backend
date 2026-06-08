package request

type AuthRequest struct {
	Phone    string `json:"phone"`
	Password string `json:"password"`
}

type RegisterRequest struct {
	PhoneHash  string   `json:"phone_hash" gorm:"column:phone_hash"`
	RoleID     int      `json:"role_id" gorm:"column:role_id"`
	Name       string   `json:"name" gorm:"column:name"`
	Gender     int      `json:"gender" gorm:"column:gender"`
	BaseSalary string   `json:"base_salary" gorm:"column:base_salary"`
	CompanyID  int      `json:"company_id" gorm:"column:company_id"`
	CheckIn1   []string `json:"check_in1" gorm:"column:check_in1"`
	CheckOut1  []string `json:"check_out1" gorm:"column:check_out1"`
	CheckIn2   []string `json:"check_in2" gorm:"column:check_in2"`
	CheckOut2  []string `json:"check_out2" gorm:"column:check_out2"`
	IsHalft    []bool   `json:"is_halft" gorm:"column:is_halft"`
	Day        []int    `json:"day" gorm:"column:day"`
	IsDayoff   []bool   `json:"is_dayoff" gorm:"column:is_dayoff"`
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

type NewPasswordRequest struct {
	NewPassword string `json:"new_password"`
}

type UserRequestUpdate struct {
	PhoneHash  *string `json:"phone_hash"`
	RoleID     *int    `json:"role_id"`
	Name       *string `json:"name"`
	Gender     *int    `json:"gender"`
	BaseSalary *string `json:"base_salary"`
}
