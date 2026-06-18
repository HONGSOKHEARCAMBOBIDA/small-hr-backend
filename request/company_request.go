package request

type CompanyRequestCreate struct {
	Name             string `json:"name" bind:"required"`
	MapLink          string `json:"map_link"`
	Radius           int    `json:"radius" bind:"required"`
	GroupLink        string `json:"group_link"`
	BotToken         string `json:"bot_token"`
	Currency         string `json:"currency"`
	LatePenalty      int    `json:"late_penalty" gorm:"column:late_penalty"`
	LeftEarlyPenalty int    `json:"left_early_penalty" gorm:"column:left_early_penalty"`
	CanScanOutsize   int    `json:"can_scan_outsize" gorm:"column:can_scan_outsize"`
}

type CompanyRequesUpdate struct {
	Name             *string `json:"name"`
	MapLink          *string `json:"map_link"`
	Latitude         *string `json:"latitude"`
	Longitude        *string `json:"longitude"`
	Radius           *int    `json:"radius"`
	Currency         *string `json:"currency"`
	LatePenalty      *int    `json:"late_penalty" gorm:"column:late_penalty"`
	LeftEarlyPenalty *int    `json:"left_early_penalty" gorm:"column:left_early_penalty"`
	CanScanOutsize   *int    `json:"can_scan_outsize" gorm:"column:can_scan_outsize"`
}

type CompanyRequestUpdateTelegram struct {
	BotToken  *string `json:"bot_token"`
	GroupLink *string `json:"group_link"`
}
