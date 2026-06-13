package service

import (
	"context"
	"errors"
	"mysql/config"
	"mysql/model"
	"mysql/request"

	"gorm.io/gorm"
)

type CompanyService interface {
	GetCompany(id int, ctx context.Context, pf request.Pagination) ([]model.Company, *model.PaginationMetadata, error)
	CreateCompany(ctx context.Context, input request.CompanyRequestCreate) error
	UpdateCompany(ctx context.Context, id int, input request.CompanyRequesUpdate) error
	ChangeStatusCompany(ctx context.Context, id int) error
}

type companyservice struct {
	db *gorm.DB
}

func NewCompanyService() CompanyService {
	return &companyservice{
		db: config.DB,
	}
}

func (s *companyservice) GetCompany(id int, ctx context.Context, pf request.Pagination) ([]model.Company, *model.PaginationMetadata, error) {
	var Company []model.Company
	var user model.User
	if err := s.db.Preload("Role").First(&user, id).Error; err != nil {
		return nil, nil, err
	}
	var totalCount int64
	offset := (pf.Page - 1) * pf.PageSize
	query := s.db.WithContext(ctx).Table("company AS c").
		Select(`
		c.id AS id,
		c.name AS name,
		c.is_active AS is_active,
		c.latitude AS latitude,
		c.longitude AS longitude,
		c.radius AS radius,
		c.bot_token AS bot_token,
		c.group_chatID AS group_chatID,
		c.currency AS currency,
		c.late_penalty AS late_penalty,
		c.left_early_penalty AS left_early_penalty
	`)

	if user.Role.Level < 7 {
		query = query.Where("c.id = ?", user.CompanyID)
	}

	if err := query.Count(&totalCount).Error; err != nil {
		return nil, nil, err
	}

	if err := query.Limit(pf.PageSize).Offset(offset).Scan(&Company).Error; err != nil {
		return nil, nil, err
	}

	totalPages := totalCount / int64(pf.PageSize)

	if int(totalCount)%pf.PageSize != 0 {
		totalPages++
	}

	metadata := &model.PaginationMetadata{
		Page:       pf.Page,
		PageSize:   pf.PageSize,
		TotalCount: totalCount,
		TotalPages: int(totalPages),
	}

	return Company, metadata, nil
}

func (s *companyservice) CreateCompany(ctx context.Context, input request.CompanyRequestCreate) error {
	tx := s.db.Begin()
	if tx.Error != nil {
		return tx.Error
	}
	newCompany := model.Company{
		Name:             input.Name,
		Latitude:         input.Latitude,
		Longitude:        input.Longitude,
		Radius:           input.Radius,
		Isactive:         true,
		BotToken:         &input.BotToken,
		GroupChatID:      &input.GroupChatID,
		Currency:         input.Currency,
		LatePenalty:      input.LatePenalty,
		LeftEarlyPenalty: input.LeftEarlyPenalty,
	}

	if err := tx.WithContext(ctx).
		Create(&newCompany).Error; err != nil {
		tx.Rollback()
		return err
	}
	return tx.Commit().Error
}

func (s *companyservice) UpdateCompany(ctx context.Context, id int, input request.CompanyRequesUpdate) error {
	updates := map[string]interface{}{}

	if input.Name != nil {
		updates["name"] = *input.Name
	}
	if input.Latitude != nil {
		updates["latitude"] = *input.Latitude
	}
	if input.Longitude != nil {
		updates["longitude"] = *input.Longitude
	}
	if input.Radius != nil {
		updates["radius"] = *input.Radius
	}
	if input.BotToken != nil {
		updates["bot_token"] = *input.BotToken
	}
	if input.GroupChatID != nil {
		updates["group_chatID"] = *input.GroupChatID
	}
	if input.Currency != nil {
		updates["currency"] = *input.Currency
	}
	if input.LatePenalty != nil {
		updates["late_penalty"] = *input.LatePenalty
	}
	if input.LeftEarlyPenalty != nil {
		updates["left_early_penalty"] = *input.LeftEarlyPenalty
	}
	if len(updates) == 0 {
		return errors.New(" no field to update")
	}
	result := s.db.WithContext(ctx).Model(&model.Company{}).Where("id =?", id).Updates(updates)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

func (s *companyservice) ChangeStatusCompany(ctx context.Context, id int) error {
	result := s.db.WithContext(ctx).Model(&model.Company{}).Where("id =?", id).Update("is_active", gorm.Expr("NOT is_active"))
	if result.Error != nil {
		return result.Error
	}
	return nil
}
