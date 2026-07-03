package response

type LeaveRequestResponse struct {
	ID                int     `json:"id"`
	UserID            int     `json:"user_id" gorm:"column:user_id"`
	UserGender        int     `json:"gender" gorm:"column:gender"`
	UserName          string  `json:"user_name"`
	RoleName          string  `json:"role_name"`
	CompanyName       string  `json:"company_name"`
	LeaveTypeID       int     `json:"leave_type_id" gorm:"column:leave_type_id"`
	LeaveTypeCode     string  `json:"leave_type_code"`
	LeaveTypeName     string  `json:"leave_type_name"`
	LeaveTypeIsdeduct bool    `json:"leave_type_is_deduct" gorm:"column:leave_type_is_deduct"`
	StartDate         string  `json:"start_date" gorm:"column:start_date"`
	EndDate           string  `json:"end_date" gorm:"column:end_date"`
	BackToWorkDate    string  `json:"back_to_work_date" gorm:"column:back_to_work_date"`
	TotalDay          float64 `json:"total_day" gorm:"column:total_day"`
	DeductTypeID      int     `json:"deduct_type_id" gorm:"column:deduct_type_id"`
	DeductTypeCode    string  `json:"deduct_type_code"`
	DeductTypeName    string  `json:"deduct_type_name"`
	Reason            string  `json:"reason" gorm:"column:reason"`
	Status            int     `json:"status" gorm:"column:status"`
	StatusString      string  `json:"status_string"`
	ApproveBy         int     `json:"approve_by" gorm:"column:approve_by"`
	ApproveByName     string  `json:"approve_by_name"`
	ApproveAt         string  `json:"approved_at" gorm:"column:approved_at"`
}
