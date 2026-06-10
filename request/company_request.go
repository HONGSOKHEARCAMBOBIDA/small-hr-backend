package request

type CompanyRequestCreate struct {
	Name             string `json:"name" bind:"required"`
	Latitude         string `json:"latitude" bind:"required"`
	Longitude        string `json:"longitude" bind:"required"`
	Radius           string `json:"radius" bind:"required"`
	BotToken         string `json:"bot_token"`
	GroupChatID      string `json:"group_chatID"`
	Currency         string `json:"currency"`
	LatePenalty      string `json:"late_penalty" gorm:"column:late_penalty"`
	LeftEarlyPenalty string `json:"left_early_penalty" gorm:"column:left_early_penalty"`
}

type CompanyRequesUpdate struct {
	Name             *string `json:"name"`
	Latitude         *string `json:"latitude"`
	Longitude        *string `json:"longitude"`
	Radius           *string `json:"radius"`
	BotToken         *string `json:"bot_token"`
	GroupChatID      *string `json:"group_chatID"`
	Currency         *string `json:"currency"`
	LatePenalty      *string `json:"late_penalty" gorm:"column:late_penalty"`
	LeftEarlyPenalty *string `json:"left_early_penalty" gorm:"column:left_early_penalty"`
}
