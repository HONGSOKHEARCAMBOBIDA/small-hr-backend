package response

import "mysql/model/base"

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

type CountUnPaidAttendance struct {
	AttendanceID int `json:"attendance_id"`
}

type PayrollResponse struct {
	base.ModelBase
	UserName       string `json:"user_name"`
	UserGender     int    `json:"user_gender"`
	RoleName       string `json:"role_name"`
	CompanyName    string `json:"company_name"`
	BasicSalary    string `json:"basic_salary" gorm:"column:basic_salary"`
	HalfSalary     string `json:"half_salary" gorm:"column:half_salary"`
	Other          string `json:"other" gorm:"column:other"`
	TotalWorkDay   int    `json:"total_work_day" gorm:"column:total_work_day"`
	TotalDeduction string `json:"total_deduction" gorm:"column:total_deduction"`
	NetSalary      string `json:"net_salary" gorm:"column:net_salary"`
	Currency       string `json:"currency"`
	PayrollType    int    `json:"payroll_type" gorm:"column:payroll_type"`
	PayrollDate    string `json:"payroll_date" gorm:"column:payroll_date"`
	Status         int    `json:"status" gorm:"column:status"`
	Note           string `json:"note" gorm:"column:note"`
}
