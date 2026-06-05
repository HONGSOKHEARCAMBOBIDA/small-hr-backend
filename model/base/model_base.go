package base

type ModelBase struct {
	ID int `json:"id" gorm:"column:id"`
}
