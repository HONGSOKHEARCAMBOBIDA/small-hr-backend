package response

type LeaveTypeResponse struct {
	ID          int    `json:"id"`
	CompanyID   int    `json:"company_id" gorm:"column:company_id"`
	CompanyName string `json:"company_name" gorm:"column:company_name"`
	Code        string `json:"code" gorm:"column:code"`
	Name        string `json:"name" gorm:"column:name"`
	Isactive    bool   `json:"is_active" gorm:"column:is_active"`
	IsDeduct    bool   `json:"is_deduct" gorm:"column:is_deduct"`
}
