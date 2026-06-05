package utils

import "github.com/google/uuid"

func GenerateQRToken() string {
	return uuid.New().String()
}
