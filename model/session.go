package model

import (
	"mysql/model/base"
	"time"
)

type Session struct {
	base.ModelBase
	UserID       uint      `json:"user_id" gorm:"not null;index"`
	RefreshToken string    `json:"-" gorm:"size:500;uniqueIndex;not null"` // hide from API
	TokenPrefix  string    `json:"token_prefix" gorm:"size:16;index;not null"`
	ExpiresAt    time.Time `json:"expires_at" gorm:"not null"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}
