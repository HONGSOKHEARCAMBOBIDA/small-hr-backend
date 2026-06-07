package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"mysql/config"
	"mysql/helper"
	"mysql/model"
	"mysql/request"
	"mysql/response"
	"mysql/utils"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type AuthService interface {
	Login(input request.AuthRequest, c *gin.Context) (*response.AuthResponse, error)
	RefreshToken(input request.RefreshTokenRequest, c *gin.Context) (*response.AuthResponse, error)
	Register(ctx context.Context, input request.RegisterRequest, c *gin.Context) error
	GetUser(ctx context.Context, id int, pf request.Pagination, filter map[string]string) ([]response.UserResponse, *model.PaginationMetadata, error)
	ToggleUserStatus(ctx context.Context, userID int, id int) error
}

type authservice struct {
	db *gorm.DB
}

func NewAuthService() AuthService {
	return &authservice{
		db: config.DB,
	}
}

func (s *authservice) Login(input request.AuthRequest, c *gin.Context) (*response.AuthResponse, error) {
	key := "login_attempt:" + input.Phone
	attempts, _ := utils.Redis.Get(utils.Ctx, key).Int()
	if attempts >= 5 {
		return nil, errors.New("អ្នកព្យាយាមចូលច្រើនពេក សូមព្យាយាមម្តងទៀតក្រោយ 10 នាទី")
	}
	var user model.User
	if err := s.db.
		Where("phone_hash = ? AND is_active = 1",
			input.Phone).
		First(&user).Error; err != nil {
		return nil, errors.New("ព័ត៌មានមិនត្រឹមត្រូវ ឬ អ្នកប្រើប្រាស់ត្រូវបានបិទគណនី")
	}

	if err := bcrypt.CompareHashAndPassword(
		[]byte(user.PasswordHash),
		[]byte(input.Password),
	); err != nil {
		utils.Redis.Incr(utils.Ctx, key)
		utils.Redis.Expire(utils.Ctx, key, 10*time.Minute)
		return nil, errors.New("ព័ត៌មានមិនត្រឹមត្រូវ")
	}
	utils.Redis.Del(utils.Ctx, key)

	accessExpiry := time.Now().Add(60 * time.Minute)
	claims := jwt.MapClaims{
		"user_id": user.ID,
		"phone":   user.PhoneHash,
		"role_id": user.RoleID,
		"exp":     accessExpiry.Unix(),
	}
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	accessTokenStr, err := accessToken.SignedString(utils.Jwtkey)
	if err != nil {
		return nil, fmt.Errorf("failed to sign access token: %w", err)
	}

	refreshTokenBytes := make([]byte, 32)
	if _, err := rand.Read(refreshTokenBytes); err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}
	refreshTokenStr := hex.EncodeToString(refreshTokenBytes)
	tokenPrefix := refreshTokenStr[:16]
	hashedRefresh, err := bcrypt.GenerateFromPassword([]byte(refreshTokenStr), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash refresh token: %w", err)
	}
	var sessionCount int64
	s.db.Model(&model.Session{}).
		Where("user_id = ? AND expires_at > ?", user.ID, time.Now()).
		Count(&sessionCount)

	if sessionCount >= 5 {
		s.db.Where("user_id = ? AND expires_at > ?", user.ID, time.Now()).
			Order("created_at ASC").
			Limit(1).
			Delete(&model.Session{})
	}

	session := model.Session{
		UserID:       uint(user.ID),
		RefreshToken: string(hashedRefresh),
		TokenPrefix:  tokenPrefix,
		ExpiresAt:    time.Now().Add(30 * 24 * time.Hour),
	}
	if err := s.db.Create(&session).Error; err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	resp := &response.AuthResponse{
		ID:           user.ID,
		Name:         user.PhoneHash,
		AccessToken:  accessTokenStr,
		RefreshToken: refreshTokenStr,
	}

	return resp, nil
}

func (s *authservice) RefreshToken(input request.RefreshTokenRequest, c *gin.Context) (*response.AuthResponse, error) {
	if len(input.RefreshToken) < 16 {
		return nil, errors.New("Invalid refresh token")
	}
	prefix := input.RefreshToken[:16]
	var session model.Session
	err := s.db.Where("token_prefix = ? AND expires_at > ?",
		prefix, time.Now()).
		First(&session).Error

	if err != nil {
		return nil, errors.New("Invalid or expired refresh token")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(session.RefreshToken), []byte(input.RefreshToken)); err != nil {
		return nil, errors.New("Invalid or expired refresh token")
	}

	newRefreshBytes := make([]byte, 32)
	rand.Read(newRefreshBytes)
	newRefreshStr := hex.EncodeToString(newRefreshBytes)
	newHash, _ := bcrypt.GenerateFromPassword([]byte(newRefreshStr), bcrypt.DefaultCost)
	newPrefix := newRefreshStr[:16]

	s.db.Model(&session).Updates(model.Session{
		RefreshToken: string(newHash),
		TokenPrefix:  newPrefix,
		ExpiresAt:    time.Now().Add(30 * 24 * time.Hour),
	})

	var user model.User
	s.db.First(&user, session.UserID)

	accessExpiry := time.Now().Add(60 * time.Minute)
	claims := jwt.MapClaims{
		"user_id": user.ID,
		"role_id": user.RoleID,
		"exp":     accessExpiry.Unix(),
	}
	accessToken, _ := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(utils.Jwtkey)

	return &response.AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: newRefreshStr,
	}, nil
}

func (s *authservice) Register(ctx context.Context, input request.RegisterRequest, c *gin.Context) error {

	if len(input.Day) != len(input.CheckIn1) ||
		len(input.Day) != len(input.CheckOut1) ||
		len(input.Day) != len(input.IsHalft) ||
		len(input.Day) != len(input.IsDayoff) {
		return errors.New("shift fields must have equal length")
	}

	passwordHash := utils.HasPassword("12345678")
	qrToken := utils.GenerateQRToken()

	tx := s.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return tx.Error
	}

	committed := false
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		} else if !committed {
			tx.Rollback()
		}
	}()

	user := model.User{
		PhoneHash:    input.PhoneHash,
		PasswordHash: passwordHash,
		RoleID:       input.RoleID,
		IsActive:     true,
		Name:         input.Name,
		Gender:       input.Gender,
		BaseSalary:   input.BaseSalary,
		CompanyID:    input.CompanyID,
		QrToken:      qrToken,
		IsVerify:     false,
	}
	if err := tx.Create(&user).Error; err != nil {
		return err
	}

	if len(input.Day) > 0 {
		shifts := make([]model.Shift, len(input.Day))
		for i := range input.Day {
			shifts[i] = model.Shift{
				UserID:    user.ID,
				Day:       input.Day[i],
				IsHalft:   input.IsHalft[i],
				IsDayoff:  input.IsDayoff[i],
				CheckIn1:  input.CheckIn1[i],
				CheckOut1: input.CheckOut1[i],
			}
			if i < len(input.CheckIn2) {
				shifts[i].CheckIn2 = input.CheckIn2[i]
			}
			if i < len(input.CheckOut2) {
				shifts[i].CheckOut2 = input.CheckOut2[i]
			}
		}
		if err := tx.Create(&shifts).Error; err != nil {
			return err
		}
	}

	if err := tx.Commit().Error; err != nil {
		return err
	}
	committed = true
	return nil
}

func applyAccessFilter(query *gorm.DB, db *gorm.DB, role model.Role, user model.User) *gorm.DB {
	if role.Level < 7 {
		return query.Where("u.company_id = ?", user.CompanyID)
	}
	return query
}

func applyCommonFilter(query *gorm.DB, filter map[string]string) *gorm.DB {
	for key, value := range filter {
		if value == "" {
			continue
		}
		switch key {
		case "name":
			query = query.Where("u.name LIKE ?", "%"+value+"%")
		case "company_id":
			query = query.Where("u.company_id =?", value)
		case "role_id":
			query = query.Where("u.role_id =?", value)
		}
	}
	return query
}

func (s *authservice) GetUser(ctx context.Context, id int, pf request.Pagination, filter map[string]string) ([]response.UserResponse, *model.PaginationMetadata, error) {
	var users []response.UserResponse
	var user model.User
	if err := s.db.WithContext(ctx).Preload("Role").First(&user, id).Error; err != nil {
		return nil, nil, err
	}

	offset := (pf.Page - 1) * pf.PageSize
	userquery := s.db.WithContext(ctx).Table("user u").
		Select(`
            u.id AS id,
            u.phone_hash AS phone_hash,
            r.id AS role_id,
            r.display_name AS role_name,
            u.is_active AS is_active,
            u.name AS name,
            u.gender AS gender,
            u.base_salary AS base_salary,
            c.id AS company_id,
            c.name AS company_name,
            u.qr_token AS qr_token,
            u.is_verify AS is_verify,
			s.value AS currency
        `).
		Joins("LEFT JOIN role r ON r.id = u.role_id").
		Joins("LEFT JOIN company c ON c.id = u.company_id").
		Joins("LEFT JOIN setting s ON s.key = ?", "CURRENCY")

	userquery = applyAccessFilter(userquery, s.db, user.Role, user)
	userquery = applyCommonFilter(userquery, filter)

	var totalCount int64
	countQuery := userquery.Session(&gorm.Session{})
	if err := countQuery.Count(&totalCount).Error; err != nil {
		return nil, nil, err
	}

	if err := userquery.Offset(offset).Limit(pf.PageSize).Scan(&users).Error; err != nil {
		return nil, nil, err
	}

	for i := range users {
		users[i].GenderString = helper.Gender(users[i].Gender)
	}

	if len(users) == 0 {
		return users, helper.BuildPaginationMeta(pf, totalCount), nil
	}

	userIDs := make([]int, len(users))
	for i, a := range users {
		userIDs[i] = a.ID
	}

	var shifts []response.ShiftResponse
	shiftquery := s.db.WithContext(ctx).Table("shift s").
		Select(`
            s.id AS id,
            s.user_id AS user_id,
            s.check_in1 AS check_in1,
            s.check_out1 AS check_out1,
            s.check_in2 AS check_in2,
            s.check_out2 AS check_out2,
            s.is_halft AS is_halft,
            s.day AS day,
            s.is_dayoff AS is_dayoff
        `).Where("s.user_id IN ?", userIDs)

	if err := shiftquery.Scan(&shifts).Error; err != nil {
		return nil, nil, err
	}

	for i := range shifts {
		shifts[i].DayName = helper.DayKhmer(shifts[i].Day)
	}

	shiftByUserID := make(map[int][]response.ShiftResponse, len(users))
	for _, r := range shifts {
		shiftByUserID[r.UserID] = append(shiftByUserID[r.UserID], r)
	}

	for i, a := range users {
		users[i].ShiftResponse = shiftByUserID[a.ID]
	}

	return users, helper.BuildPaginationMeta(pf, totalCount), nil
}

func (s *authservice) ToggleUserStatus(ctx context.Context, userID int, id int) error {
	result := s.db.WithContext(ctx).Model(&model.User{}).Where("id =?", id).Update("is_active", gorm.Expr("NOT is_active"))
	if result.Error != nil {
		return result.Error
	}
	return nil
}
