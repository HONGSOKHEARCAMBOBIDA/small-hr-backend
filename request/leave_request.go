package request

type LeaveRequestCreate struct {
	LeaveTypeID    int     `json:"leave_type_id" gorm:"column:leave_type_id"`
	StartDate      string  `json:"start_date" gorm:"column:start_date"`
	EndDate        string  `json:"end_date" gorm:"column:end_date"`
	BackToWorkDate string  `json:"back_to_work_date" gorm:"column:back_to_work_date"`
	TotalDay       float64 `json:"total_day" gorm:"column:total_day"`
	DeductTypeID   int     `json:"deduct_type_id" gorm:"column:deduct_type_id"`
	Reason         *string `json:"reason" gorm:"column:reason"`
	ApproveBy      int     `json:"approve_by" gorm:"column:approve_by"`
}

type LeaveRequestUpdate struct {
	LeaveTypeID    *int     `json:"leave_type_id" gorm:"column:leave_type_id"`
	StartDate      *string  `json:"start_date" gorm:"column:start_date"`
	EndDate        *string  `json:"end_date" gorm:"column:end_date"`
	BackToWorkDate *string  `json:"back_to_work_date" gorm:"column:back_to_work_date"`
	TotalDay       *float64 `json:"total_day" gorm:"column:total_day"`
	DeductTypeID   *int     `json:"deduct_type_id" gorm:"column:deduct_type_id"`
	Reason         *string  `json:"reason" gorm:"column:reason"`
	ApproveBy      *int     `json:"approve_by" gorm:"column:approve_by"`
}

type LeaveRequestUpdateStatus struct {
	Status *int `json:"status" gorm:"column:status"`
}
