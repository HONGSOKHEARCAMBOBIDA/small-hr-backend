package model

import "mysql/model/base"

type Company struct {
	base.ModelBase
	Name             string  `json:"name"`
	Isactive         bool    `json:"is_active" gorm:"column:is_active"`
	Latitude         string  `json:"latitude"`
	Longitude        string  `json:"longitude"`
	Radius           string  `json:"radius" gorm:"column:radius"`
	BotToken         *string `json:"bot_token" gorm:"column:bot_token"`
	GroupChatID      *string `json:"group_chatID" gorm:"column:group_chatID"`
	Currency         string  `json:"currency" gorm:"column:currency"`
	LatePenalty      string  `json:"late_penalty" gorm:"column:late_penalty"`
	LeftEarlyPenalty string  `json:"left_early_penalty" gorm:"column:left_early_penalty"`
	CanScanOutsize   int     `json:"can_scan_outsize" gorm:"column:can_scan_outsize"`
	Color            string  `json:"color" gorm:"column:color;default:'#000000'"`
	TotalWorkDay     int     `json:"total_work_day" gorm:"column:total_work_day"`
}

func (Company) TableName() string {
	return "company"
}
