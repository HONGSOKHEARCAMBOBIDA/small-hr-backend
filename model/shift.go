package model

import (
	"mysql/model/base"
)

type Shift struct {
	base.ModelBase
	UserID    int     `json:"user_id" gorm:"column:user_id"`
	CheckIn1  *string `json:"check_in1" gorm:"column:check_in1"`
	CheckOut1 *string `json:"check_out1" gorm:"column:check_out1"`
	CheckIn2  *string `json:"check_in2" gorm:"column:check_in2"`
	CheckOut2 *string `json:"check_out2" gorm:"column:check_out2"`
	ShiftType int     `json:"shift_type" gorm:"column:shift_type"` // 1=Full, 2=Morning only, 3=Evening only
	Day       int     `json:"day" gorm:"column:day"`
	IsDayoff  bool    `json:"is_dayoff" gorm:"column:is_dayoff"`
}

func (Shift) TableName() string {
	return "shift"
}
