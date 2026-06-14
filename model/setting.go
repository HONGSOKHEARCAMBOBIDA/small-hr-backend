package model

import "mysql/model/base"

type Setting struct {
	base.ModelBase
	Key   string `json:"key" gorm:"column:key"`
	Value string `json:"value" gorm:"column:value"`
}

func (Setting) TableName() string {
	return "setting"
}
