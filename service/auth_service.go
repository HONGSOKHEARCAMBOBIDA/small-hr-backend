package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"mysql/config"
	"mysql/helper"
	"mysql/model"
	"mysql/request"
	"mysql/response"
	"mysql/utils"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type AuthService interface {
	Login(input request.AuthRequest, c *gin.Context) (*response.AuthResponse, error)
	LoginByQr(input request.LoginQrRequest, c *gin.Context) (*response.AuthResponse, error)
	RefreshToken(input request.RefreshTokenRequest, c *gin.Context) (*response.AuthResponse, error)
	Register(ctx context.Context, input request.RegisterRequest, c *gin.Context, userID int) error
	GetUser(ctx context.Context, id int, pf request.Pagination, filter map[string]string) ([]response.UserResponse, *model.PaginationMetadata, error)
	ToggleUserStatus(ctx context.Context, id int) error
	ChangePassword(ctx context.Context, userID int, input request.NewPasswordRequest) error
	UpdateUser(ctx context.Context, input request.UserRequestUpdate, id int) error
	CountUser(ctx context.Context, id int) (response.UserCount, error)
	GetRole(ctx context.Context, id int) ([]model.Role, error)
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
	phonehash := helper.HashPhone(input.Phone)
	var user model.User
	if err := s.db.Select("id, phone_hash, password_hash, role_id, is_active, name").
		Where("phone_hash = ? AND is_active = 1", phonehash).
		First(&user).Error; err != nil {
		return nil, errors.New("ព័ត៌មានមិនត្រឹមត្រូវ ឬ អ្នកប្រើប្រាស់ត្រូវបានបិទគណនី")
	}

	var settings []model.Setting
	if err := s.db.Where("`key` IN ?", []string{
		"ACCESS_TOKEN_EXPIRE_HOURS",
		"REFRESH_TOKEN_EXPIRE_DAYS",
	}).Find(&settings).Error; err != nil {
		return nil, errors.New("Setting Not Found")
	}

	settingMap := make(map[string]string)
	for _, s := range settings {
		settingMap[s.Key] = s.Value
	}

	accesstoken, err := strconv.Atoi(settingMap["ACCESS_TOKEN_EXPIRE_HOURS"])
	if err != nil {
		return nil, errors.New("Bad request")
	}
	refreshtoken, err := strconv.Atoi(settingMap["REFRESH_TOKEN_EXPIRE_DAYS"])
	if err != nil {
		return nil, errors.New("Bad request")
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

	var permissions []model.Permission
	if err := s.db.Table("permission p").
		Select("p.id AS id, p.name AS name").
		Joins("JOIN role_permission rhp ON rhp.permission_id = p.id").
		Where("rhp.role_id = ? AND p.name IN ?", user.RoleID, []string{
			"add.payroll", "add.backup", "view.backup", "view.download.backup", "delete.backup",
		}).
		Scan(&permissions).Error; err != nil {
		return nil, err
	}

	accessExpiry := time.Now().Add(time.Duration(accesstoken) * time.Hour)
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

	if err := s.db.Where("user_id = ?", user.ID).Delete(&model.Session{}).Error; err != nil {
		log.Printf(err.Error())
		return nil, fmt.Errorf("failed to delete session")
	}

	session := model.Session{
		UserID:       uint(user.ID),
		RefreshToken: string(hashedRefresh),
		TokenPrefix:  tokenPrefix,
		ExpiresAt:    time.Now().Add(time.Duration(refreshtoken) * 24 * time.Hour),
	}
	if err := s.db.Create(&session).Error; err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	resp := &response.AuthResponse{
		ID:           user.ID,
		Name:         user.Name,
		AccessToken:  accessTokenStr,
		RefreshToken: refreshTokenStr,
		Permissions:  permissions,
	}

	return resp, nil
}

func (s *authservice) LoginByQr(input request.LoginQrRequest, c *gin.Context) (*response.AuthResponse, error) {

	qrtokenhash := helper.HashQrtoken(input.QrToken)
	var user model.User
	if err := s.db.Select("id,qr_token,is_active,role_id,name").
		Where("qr_token = ? AND is_active = 1", qrtokenhash).
		First(&user).Error; err != nil {
		return nil, errors.New("ព័ត៌មានមិនត្រឹមត្រូវ ឬ អ្នកប្រើប្រាស់ត្រូវបានបិទគណនី")
	}

	var permissions []model.Permission
	if err := s.db.Table("permission p").
		Select("p.id AS id, p.name AS name").
		Joins("JOIN role_permission rhp ON rhp.permission_id = p.id").
		Where("rhp.role_id = ? AND p.name IN ?", user.RoleID, []string{
			"add.payroll",
		}).
		Scan(&permissions).Error; err != nil {
		return nil, err
	}

	var settings []model.Setting
	if err := s.db.Where("`key` IN ?", []string{
		"ACCESS_TOKEN_EXPIRE_HOURS",
		"REFRESH_TOKEN_EXPIRE_DAYS",
	}).Find(&settings).Error; err != nil {
		return nil, errors.New("Setting Not Found")
	}

	settingMap := make(map[string]string)
	for _, s := range settings {
		settingMap[s.Key] = s.Value
	}

	accesstoken, err := strconv.Atoi(settingMap["ACCESS_TOKEN_EXPIRE_HOURS"])
	if err != nil {
		return nil, errors.New("Bad request")
	}
	refreshtoken, err := strconv.Atoi(settingMap["REFRESH_TOKEN_EXPIRE_DAYS"])
	if err != nil {
		return nil, errors.New("Bad request")
	}

	newQrToken := utils.GenerateQRToken()
	qrTokenHash := helper.HashQrtoken(newQrToken)
	qrTokenEncript, err := helper.EncryptQRTOKEN(newQrToken)
	if err != nil {
		return nil, err
	}
	if err := s.db.Model(&model.User{}).
		Where("id = ?", user.ID).
		Updates(map[string]interface{}{
			"qr_token":         qrTokenHash,
			"qr_token_encript": qrTokenEncript,
		}).Error; err != nil {
		return nil, fmt.Errorf("failed to rotate qr token: %w", err)
	}

	accessExpiry := time.Now().Add(time.Duration(accesstoken) * time.Hour)
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
		ExpiresAt:    time.Now().Add(time.Duration(refreshtoken) * 24 * time.Hour),
	}
	if err := s.db.Create(&session).Error; err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	resp := &response.AuthResponse{
		ID:           user.ID,
		Name:         user.Name,
		AccessToken:  accessTokenStr,
		RefreshToken: refreshTokenStr,
		Permissions:  permissions,
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

func (s *authservice) Register(ctx context.Context, input request.RegisterRequest, c *gin.Context, userID int) error {

	if len(input.Day) != len(input.CheckIn1) ||
		len(input.Day) != len(input.CheckOut1) ||
		len(input.Day) != len(input.ShiftType) ||
		len(input.Day) != len(input.IsDayoff) {
		return errors.New("shift fields must have equal length")
	}

	var usrlog model.User
	if err := s.db.WithContext(ctx).First(&usrlog, userID).Error; err != nil {
		return err
	}
	PhoneEncript, err := helper.EncryptPhone(input.PhoneHash)
	if err != nil {
		return err
	}
	passwordHash := utils.HasPassword("12345678")
	qrToken := utils.GenerateQRToken()
	qrTokenHash := helper.HashQrtoken(qrToken)
	qrTokenEncript, err := helper.EncryptQRTOKEN(qrToken)
	if err != nil {
		return err
	}
	if err != nil {
		return err
	}
	companyID := input.CompanyID
	if companyID == 0 {
		companyID = usrlog.CompanyID
	}
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
		PhoneHash:      helper.HashPhone(input.PhoneHash),
		PhoneEncript:   PhoneEncript,
		PasswordHash:   passwordHash,
		RoleID:         input.RoleID,
		IsActive:       true,
		Name:           input.Name,
		Gender:         input.Gender,
		BaseSalary:     input.BaseSalary,
		CompanyID:      companyID,
		QrToken:        qrTokenHash,
		IsVerify:       false,
		QrTokenEncript: qrTokenEncript,
	}
	if err := tx.Create(&user).Error; err != nil {
		return err
	}

	for i, day := range input.Day {
		shift := model.Shift{
			UserID:    user.ID,
			Day:       day,
			ShiftType: input.ShiftType[i],
			IsDayoff:  input.IsDayoff[i],
			CheckIn1:  input.CheckIn1[i],
			CheckOut1: input.CheckOut1[i],
			CheckIn2:  input.CheckIn2[i],
			CheckOut2: input.CheckOut2[i],
		}
		if err := tx.Create(&shift).Error; err != nil {
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
	if role.Level > 1 && role.Level < 7 {
		return query.Where("u.company_id = ?", user.CompanyID)
	} else if role.Level <= 1 {
		return query.Where("u.id =?", user.ID)
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
            u.phone_encrypted AS phone_hash,
            r.id AS role_id,
            r.display_name AS role_name,
            u.is_active AS is_active,
            u.name AS name,
            u.gender AS gender,
            u.base_salary AS base_salary,
            c.id AS company_id,
            c.name AS company_name,
            u.qr_token_encript AS qr_token,
            u.is_verify AS is_verify,
			c.currency AS currency
        `).
		Joins("LEFT JOIN role r ON r.id = u.role_id").
		Joins("LEFT JOIN company c ON c.id = u.company_id")

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
		decrypted, err := helper.DecryptSalary(users[i].BaseSalary)
		if err != nil {
			log.Printf(err.Error())
			return nil, nil, err
		}
		users[i].BaseSalary = decrypted
		phonedecript, err := helper.DecryptPhone(users[i].PhoneHash)
		if err != nil {
			return nil, nil, err
		}
		users[i].PhoneHash = phonedecript
		qrtokendecript, err := helper.DecryptQRTOKEN(users[i].QrToken)
		if err != nil {
			return nil, nil, err
		}
		users[i].QrToken = qrtokendecript
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
            s.shift_type AS shift_type,
            s.day AS day,
            s.is_dayoff AS is_dayoff
        `).Where("s.user_id IN ?", userIDs)

	if err := shiftquery.Scan(&shifts).Error; err != nil {
		return nil, nil, err
	}

	for i := range shifts {
		shifts[i].DayName = helper.DayKhmer(shifts[i].Day)
		shifts[i].ShiftTypeString = helper.ShiftType(shifts[i].ShiftType)
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

func (s *authservice) ToggleUserStatus(ctx context.Context, id int) error {
	result := s.db.WithContext(ctx).Model(&model.User{}).Where("id =?", id).Update("is_active", gorm.Expr("NOT is_active"))
	if result.Error != nil {
		return result.Error
	}
	return nil
}

func (s *authservice) ChangePassword(ctx context.Context, userID int, input request.NewPasswordRequest) error {
	var user model.User
	if err := s.db.WithContext(ctx).First(&user, userID).Error; err != nil {
		return err
	}

	hash := utils.HasPassword(input.NewPassword)
	if err := s.db.WithContext(ctx).Model(&user).Update("password_hash", hash).Error; err != nil {
		return err
	}
	return nil

}

func (s *authservice) UpdateUser(ctx context.Context, input request.UserRequestUpdate, id int) error {
	updates := map[string]interface{}{}

	if input.PhoneHash != nil {
		updates["phone_hash"] = helper.HashPhone(*input.PhoneHash)
		PhoneEncript, err := helper.EncryptPhone(*input.PhoneHash)
		if err != nil {
			return err
		}
		updates["phone_encrypted"] = PhoneEncript
		// updates["qr_token"] = helper.HashQrtoken(*input.PhoneHash)
		// qrencript, err := helper.EncryptQRTOKEN(*input.PhoneHash)
		// if err != nil {
		// 	return err
		// }
		// updates["qr_token_encript"] = qrencript
	}
	if input.RoleID != nil {
		updates["role_id"] = *input.RoleID
	}
	if input.Name != nil {
		updates["name"] = *input.Name
	}
	if input.Gender != nil {
		updates["gender"] = *input.Gender
	}
	if input.BaseSalary != nil {
		encrypted, err := helper.EncryptSalary(*input.BaseSalary)
		if err != nil {
			return err
		}
		updates["base_salary"] = encrypted
	}
	result := s.db.WithContext(ctx).Model(&model.User{}).Where("id =?", id).Updates(updates)
	if result.Error != nil {
		slog.Error("failed to create payroll", "error", result.Error)
		return result.Error

	}
	return nil
}

func (s *authservice) CountUser(ctx context.Context, id int) (response.UserCount, error) {
	var countUser response.UserCount
	var user model.User
	if err := s.db.WithContext(ctx).Preload("Role").First(&user, id).Error; err != nil {
		return response.UserCount{}, err
	}

	userQuery := s.db.WithContext(ctx).
		Table("user u").
		Select(`
            COUNT(DISTINCT CASE WHEN u.is_active = '1' THEN u.id END) AS total
        `)

	userQuery = helper.ApplyAccessFilter(userQuery, s.db, user.Role, user)
	if err := userQuery.Scan(&countUser).Error; err != nil {
		return response.UserCount{}, err
	}
	return countUser, nil
}

func (s *authservice) GetRole(ctx context.Context, id int) ([]model.Role, error) {
	var role []model.Role
	var user model.User
	if err := s.db.WithContext(ctx).Preload("Role").First(&user, id).Error; err != nil {
		return nil, err
	}
	roleQuery := s.db.WithContext(ctx).Table("role r").
		Select(`
		 r.id AS id,
		 r.display_name AS display_name
	`)
	roleQuery = helper.ApplyAccessGetRole(roleQuery, s.db, user.Role, user)

	if err := roleQuery.Scan(&role).Error; err != nil {
		return nil, err
	}
	return role, nil
}
