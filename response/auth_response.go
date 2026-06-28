package response

import "mysql/model"

type AuthResponse struct {
	ID           int                `json:"id"`
	Name         string             `json:"name"`
	AccessToken  string             `json:"access_token"`
	RefreshToken string             `json:"refresh_token"`
	Permissions  []model.Permission `json:"permissions"`
}

type UserDataResponse struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}
