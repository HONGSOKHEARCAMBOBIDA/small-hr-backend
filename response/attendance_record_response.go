package response

type AttendanceRecordResponse struct {
	ID                 int    `json:"id"`
	AttendanceID       int    `json:"attendance_id"`
	Day                int    `json:"day"`
	DayString          string `json:"day_string"`
	AttendanceType     int    `json:"attendance_type"`
	AttendanceTypeName string `json:"attendance_type_name"`
	Reason             string `json:"reason"`
	CheckTime          string `json:"check_time"`
	Type               int    `json:"type"`
	TypeString         string `json:"type_string"`
	Inzone             int    `json:"inzone"`
	Latitude           string `json:"latitdude" gorm:"column:latitdude"`
	Longitude          string `json:"longitude" gorm:"column:longitude"`
	ScheduledTime      string `json:"scheduled_time"`
	TimeDiff           string `json:"time_diff"`
	CheckIn1           string `json:"check_in1" gorm:"column:check_in1"`
	CheckOut1          string `json:"check_out1" gorm:"column:check_out1"`
	CheckIn2           string `json:"check_in2" gorm:"column:check_in2"`
	CheckOut2          string `json:"check_out2" gorm:"column:check_out2"`
}
