package request

type CompanyRequestCreate struct {
	Name             string `json:"name" bind:"required"`
	MapLink          string `json:"map_link"`
	Radius           string `json:"radius" bind:"required"`
	GroupLink        string `json:"group_link"`
	BotToken         string `json:"bot_token"`
	Currency         string `json:"currency"`
	LatePenalty      string `json:"late_penalty" gorm:"column:late_penalty"`
	LeftEarlyPenalty string `json:"left_early_penalty" gorm:"column:left_early_penalty"`
	CanScanOutsize   int    `json:"can_scan_outsize" gorm:"column:can_scan_outsize"`
}

type CompanyRequesUpdate struct {
	Name             *string `json:"name"`
	MapLink          *string `json:"map_link"`
	Latitude         *string `json:"latitude"`
	Longitude        *string `json:"longitude"`
	Radius           *string `json:"radius"`
	Currency         *string `json:"currency"`
	LatePenalty      *string `json:"late_penalty" gorm:"column:late_penalty"`
	LeftEarlyPenalty *string `json:"left_early_penalty" gorm:"column:left_early_penalty"`
	CanScanOutsize   *int    `json:"can_scan_outsize" gorm:"column:can_scan_outsize"`
}

type CompanyRequestUpdateTelegram struct {
	BotToken  *string `json:"bot_token"`
	GroupLink *string `json:"group_link"`
}
