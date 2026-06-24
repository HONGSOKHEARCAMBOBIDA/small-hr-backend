package helper

import (
	"errors"
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

func CanManageUser(actorRole model.Role, targetRole model.Role) error {
	if targetRole.Level >= actorRole.Level {
		return errors.New("you do not have permission to manage a user with a higher role level")
	}
	return nil
}
