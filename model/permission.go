package model

type Permission struct {
	ID          uint   `gorm:"primaryKey" json:"id"`
	Name        string `gorm:"size:100;not null" json:"name"`
	DisplayName string `json:"display_name" gorm:"column:display_name"`
}

func (Permission) TableName() string {
	return "permission"
}
