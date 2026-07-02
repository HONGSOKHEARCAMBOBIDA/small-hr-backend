package request

type LeaveTypeRequestCreate struct {
	CompanyID int    `json:"company_id" gorm:"column:company_id"`
	Code      string `json:"code" gorm:"column:code"`
	Name      string `json:"name" gorm:"column:name"`
	Isactive  bool   `json:"is_active" gorm:"column:is_active"`
	IsDeduct  bool   `json:"is_deduct" gorm:"column:is_deduct"`
}

type LeaveTypeRequestUpdate struct {
	CompanyID *int    `json:"company_id" gorm:"column:company_id"`
	Code      *string `json:"code" gorm:"column:code"`
	Name      *string `json:"name" gorm:"column:name"`
	Isactive  *bool   `json:"is_active" gorm:"column:is_active"`
	IsDeduct  *bool   `json:"is_deduct" gorm:"column:is_deduct"`
}
