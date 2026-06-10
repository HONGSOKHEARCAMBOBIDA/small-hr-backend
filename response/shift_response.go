package response

import "mysql/model/base"

type ShiftResponse struct {
	base.ModelBase
	UserID          int    `json:"user_id"`
	CheckIn1        string `json:"check_in1" gorm:"column:check_in1"`
	CheckOut1       string `json:"check_out1" gorm:"column:check_out1"`
	CheckIn2        string `json:"check_in2" gorm:"column:check_in2"`
	CheckOut2       string `json:"check_out2" gorm:"column:check_out2"`
	ShiftType       int    `json:"shift_type" gorm:"column:shift_type"`
	ShiftTypeString string `json:"shift_type_string"`
	Day             int    `json:"day" gorm:"column:day"`
	DayName         string `json:"day_name"`
	IsDayoff        bool   `json:"is_dayoff" gorm:"column:is_dayoff"`
}
