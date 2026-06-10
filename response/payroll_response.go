package response

type PayrollDraftResponse struct {
	UserID                int                     `json:"user_id"`
	UserName              string                  `json:"user_name"`
	RoleName              string                  `json:"role_name"`
	BasicSalary           string                  `json:"basic_salary"`
	HalfSalary            string                  `json:"half_salary"`
	Other                 string                  `json:"other"`
	TotalLate             int                     `json:"total_late"`
	TotalLeftEarly        int                     `json:"total_left_early"`
	TotalWorkDay          int                     `json:"total_work_day"`
	TotalDeduction        string                  `json:"total_deduction"`
	NetSalary             string                  `json:"net_salary"`
	Currency              string                  `json:"currency"`
	CountUnPaidAttendance []CountUnPaidAttendance `json:"unpaid_attendance"`
}

// CountUnPaidAttendance holds one unpaid attendance event for this user.
// Type: 3 = late (ចូលធ្វើការយឺត), 4 = left early (ចេញពីធ្វើការមុនម៉ោង)
type CountUnPaidAttendance struct {
	AttendanceID int `json:"attendance_id"`
	Type         int `json:"type"`
}
