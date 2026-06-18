package service

import (
	"context"
	"errors"
	"fmt"
	"log"
	"mysql/config"
	"mysql/model"
	"mysql/request"
	"mysql/response"
	"mysql/utils"

	"gorm.io/gorm"
)

type CompanyService interface {
	GetCompany(id int, ctx context.Context, pf request.Pagination) ([]response.CompanyResponse, *model.PaginationMetadata, error)
	CreateCompany(ctx context.Context, input request.CompanyRequestCreate) error
	UpdateCompany(ctx context.Context, id int, input request.CompanyRequesUpdate) error
	ChangeStatusCompany(ctx context.Context, id int) error
	UpdateTelegram(ctx context.Context, id int, input request.CompanyRequestUpdateTelegram) error
}

type companyservice struct {
	db *gorm.DB
}

func NewCompanyService() CompanyService {
	return &companyservice{
		db: config.DB,
	}
}

func (s *companyservice) GetCompany(id int, ctx context.Context, pf request.Pagination) ([]response.CompanyResponse, *model.PaginationMetadata, error) {
	var Company []response.CompanyResponse
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
		c.group_chatID AS group_link,
		c.currency AS currency,
		c.late_penalty AS late_penalty,
		c.left_early_penalty AS left_early_penalty,
		c.can_scan_outsize AS can_scan_outsize,
		COUNT(u.id) AS user_count
	`).Joins("LEFT JOIN user AS u ON u.company_id = c.id").
		Group("c.id")

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

	for i := range Company {
		if Company[i].BotToken != nil && *Company[i].BotToken != "" {
			botTokenDecrypt, err := utils.DecryptBotToken(*Company[i].BotToken)
			if err != nil {
				return nil, nil, err
			}
			Company[i].BotToken = &botTokenDecrypt
		}

		if Company[i].GroupChatID != nil && *Company[i].GroupChatID != "" {
			chatIDDecrypt, err := utils.DecryptChatID(*Company[i].GroupChatID)
			if err != nil {
				return nil, nil, err
			}
			Company[i].GroupChatID = &chatIDDecrypt
		}
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
	chatID, err := utils.ResolveTelegramChatID(input.BotToken, input.GroupLink)
	if err != nil {
		log.Printf(err.Error())
		tx.Rollback()
		return fmt.Errorf("could not resolve group link: %w", err)
	}

	chatIDStr := fmt.Sprintf("%d", chatID)

	encryptedChatID, err := utils.EncryptChatID(chatIDStr)
	if err != nil {
		return err
	}
	encryptedBottoken, err := utils.EncryptBotToken(input.BotToken)
	if err != nil {
		return err
	}
	newCompany := model.Company{
		Name:             input.Name,
		Latitude:         input.Latitude,
		Longitude:        input.Longitude,
		Radius:           input.Radius,
		Isactive:         true,
		BotToken:         &encryptedBottoken,
		GroupChatID:      &encryptedChatID,
		Currency:         input.Currency,
		LatePenalty:      input.LatePenalty,
		LeftEarlyPenalty: input.LeftEarlyPenalty,
		CanScanOutsize:   input.CanScanOutsize,
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
	switch {
	case input.MapLink != nil && *input.MapLink != "":
		lat, lng, err := utils.ExtractLatLngFromGoogleMapsURL(*input.MapLink)
		if err != nil {
			return fmt.Errorf("invalid map_link: %w", err)
		}
		updates["latitude"] = lat
		updates["longitude"] = lng
	default:
		if input.Latitude != nil {
			updates["latitude"] = *input.Latitude
		}
		if input.Longitude != nil {
			updates["longitude"] = *input.Longitude
		}
	}
	if input.Radius != nil {
		updates["radius"] = *input.Radius
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
	if input.CanScanOutsize != nil {
		updates["can_scan_outsize"] = *input.CanScanOutsize
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

func (s *companyservice) UpdateTelegram(ctx context.Context, id int, input request.CompanyRequestUpdateTelegram) error {
	updates := map[string]interface{}{}
	if input.BotToken != nil {
		encryptedBottoken, err := utils.EncryptBotToken(*input.BotToken)
		if err != nil {
			return err
		}
		updates["bot_token"] = encryptedBottoken
	}
	if input.GroupLink != nil {
		chatID, err := utils.ResolveTelegramChatID(*input.BotToken, *input.GroupLink)
		if err != nil {
			return fmt.Errorf("could not resolve group link")
		}
		chatIDStr := fmt.Sprintf("%d", chatID)
		encryptedChatID, err := utils.EncryptChatID(chatIDStr)
		if err != nil {
			return err
		}
		updates["group_chatID"] = encryptedChatID
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
