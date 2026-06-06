package model

import (
	"mysql/model/base"
	"time"
)

type Shift struct {
	base.ModelBase
	UserID    int       `json:"user_id" gorm:"column:user_id"`
	CheckIn1  time.Time `json:"check_in1" gorm:"column:check_in1"`
	CheckOut1 time.Time `json:"check_out1" gorm:"column:check_out1"`
	CheckIn2  time.Time `json:"check_in2" gorm:"column:check_in2"`
	CheckOut2 time.Time `json:"check_out2" gorm:"column:check_out2"`
	IsHalft   bool      `json:"is_halft" gorm:"column:is_halft"`
	Day       int       `json:"day" gorm:"column:day"`
	IsDayoff  bool      `json:"is_dayoff" gorm:"column:is_dayoff"`
}

func (Shift) TableName() string {
	return "shift"
}
