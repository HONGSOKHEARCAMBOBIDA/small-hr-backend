package response

type AttendanceResponse struct {
	ID                       int                        `json:"id"`
	UserID                   int                        `json:"user_id"`
	Name                     string                     `json:"name"`
	Gender                   int                        `json:"gender"`
	GenderString             string                     `json:"gender_string"`
	CompanyID                int                        `json:"company_id"`
	CompanyName              string                     `json:"company_name"`
	RoleID                   int                        `json:"role_id"`
	RoleName                 string                     `json:"role_name"`
	CheckDate                string                     `json:"check_date"`
	Status                   string                     `json:"status"`
	AttendanceRecordResponse []AttendanceRecordResponse `json:"attendance_record" gorm:"-"`
}

type AttendanceResponseDraft struct {
	Type          int    `json:"type"`
	TypeString    string `json:"type_string"`
	ScheduledTime string `json:"scheduled_time"`
}
