package response

import "mysql/model/base"

type CompanyResponse struct {
	base.ModelBase
	Name             string  `json:"name"`
	TotalUser        int     `json:"user_count" gorm:"column:user_count"`
	Isactive         bool    `json:"is_active" gorm:"column:is_active"`
	Latitude         string  `json:"latitude"`
	Longitude        string  `json:"longitude"`
	Radius           string  `json:"radius" gorm:"column:radius"`
	BotToken         *string `json:"bot_token" gorm:"column:bot_token"`
	GroupChatID      *string `json:"group_link" gorm:"column:group_link"`
	Currency         string  `json:"currency" gorm:"column:currency"`
	LatePenalty      string  `json:"late_penalty" gorm:"column:late_penalty"`
	LeftEarlyPenalty string  `json:"left_early_penalty" gorm:"column:left_early_penalty"`
	CanScanOutsize   int     `json:"can_scan_outsize" gorm:"column:can_scan_outsize"`
	Color            string  `json:"color" gorm:"column:color"`
}

type CompanyColor struct {
	Color string `json:"color" gorm:"column:color"`
}
