package helper

import (
	"mysql/model"

	"gorm.io/gorm"
)

func ApplyAccessFilter(query, db *gorm.DB, role model.Role, user model.User) *gorm.DB {
	if role.Level < 7 {
		return query.Where("u.company_id = ?", user.CompanyID)
	}
	return query
}

func ApplyAccessGetRole(query, db *gorm.DB, role model.Role, user model.User) *gorm.DB {
	return query.Where("r.level <= ?", user.Role.Level)
}
