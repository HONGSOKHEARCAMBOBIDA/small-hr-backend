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
	RoleName                 string                     `json:"role_name" gorm:"column:role_name"`
	CheckDate                string                     `json:"check_date"`
	Status                   string                     `json:"status"`
	AttendanceRecordResponse []AttendanceRecordResponse `json:"attendance_record" gorm:"-"`
}

type AttendanceResponseDraft struct {
	Type          int    `json:"type"`
	TypeString    string `json:"type_string"`
	ScheduledTime string `json:"scheduled_time"`
}

type AttendanceResponseGenerate struct {
	ID           int    `json:"id"`
	UserID       int    `json:"user_id"`
	Name         string `json:"name"`
	Gender       int    `json:"gender"`
	GenderString string `json:"gender_string"`
	CompanyID    int    `json:"company_id"`
	CompanyName  string `json:"company_name"`
	RoleID       int    `json:"role_id"`
	RoleName     string `json:"role_name"`
	CheckDate    string `json:"check_date"`
	Status       string `json:"status"`

	CheckIn1      string `json:"check_in1"`
	CheckIn1Diff  string `json:"check_in1_diff"`
	CheckOut1     string `json:"check_out1"`
	CheckOut1Diff string `json:"check_out1_diff"`
	CheckIn2      string `json:"check_in2"`
	CheckIn2Diff  string `json:"check_in2_diff"`
	CheckOut2     string `json:"check_out2"`
	CheckOut2Diff string `json:"check_out2_diff"`

	Reason string `json:"reason"`
}
