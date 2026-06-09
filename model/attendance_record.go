package model

import "mysql/model/base"

type AttendanceRecord struct {
	base.ModelBase
	AttendanceID   int    `json:"attendance_id" gorm:"column:attendance_id"`
	ShiftID        int    `json:"shift_id" gorm:"column:shift_id"`
	AttendanceType int    `json:"attendance_type" gorm:"column:attendance_type"`
	Reason         string `json:"resean" gorm:"column:resean"`
	CheckTime      string `json:"check_time" gorm:"column:check_time"`
	Type           int    `json:"type" gorm:"column:type"`
	Inzone         bool   `json:"inzone" gorm:"column:inzone"`
}

func (AttendanceRecord) TableName() string {
	return "attendance_record"
}
