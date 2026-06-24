package helper

import (
	"mysql/model"

	"gorm.io/gorm"
)

func ManageCompanyFilter(query *gorm.DB, db *gorm.DB, user model.User) *gorm.DB {
	switch user.ManageCompany {
	case 1:
		return query.Where("c.id = ?", user.CompanyID)
	case 2:
		var companyIDs []int
		db.Model(&model.UserCompany{}).Where("user_id = ?", user.ID).Pluck("company_id", &companyIDs)
		if len(companyIDs) == 0 {
			return query.Where("1 = 0")
		}
		return query.Where("c.id IN ?", companyIDs)
	}
	return query
}
