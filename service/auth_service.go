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
	"net/http"
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
	RefreshToken(refreshToken string, c *gin.Context) (*response.AuthResponse, error)
	Register(ctx context.Context, input request.RegisterRequest, c *gin.Context, userID int) error
	GetUser(ctx context.Context, id int, pf request.Pagination, filter map[string]string) ([]response.UserResponse, *model.PaginationMetadata, error)
	ToggleUserStatus(ctx context.Context, id int, userID int) error
	ChangePassword(ctx context.Context, userID int, input request.NewPasswordRequest) error
	UpdateUser(ctx context.Context, input request.UserRequestUpdate, id int) error
	CountUser(ctx context.Context, id int) (response.UserCount, error)
	GetRole(ctx context.Context, id int) ([]model.Role, error)
	DeleteUser(ctx context.Context, id int, userIDlogin int) error
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
			"add.payroll", "add.backup", "view.backup", "view.download.backup", "delete.backup", "add.company", "edit.company", "edit.user",
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
	hashedRefresh := utils.HashToken(refreshTokenStr)
	if err := s.db.Where("user_id = ?", user.ID).Delete(&model.Session{}).Error; err != nil {
		log.Printf(err.Error())
		return nil, fmt.Errorf("failed to delete session")
	}
	refreshExpiry := time.Now().Add(time.Duration(refreshtoken) * 24 * time.Hour)
	session := model.Session{
		UserID:       uint(user.ID),
		RefreshToken: string(hashedRefresh),
		TokenPrefix:  tokenPrefix,
		ExpiresAt:    refreshExpiry,
	}
	if err := s.db.Create(&session).Error; err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	maxAge := int(time.Until(refreshExpiry).Seconds())
	c.SetSameSite(http.SameSiteStrictMode)
	c.SetCookie(
		"refresh_token",
		refreshTokenStr,
		maxAge,
		"/",
		"",   // domain: leave empty for current host, or set "yourdomain.com" explicitly
		true, // Secure - HTTPS only
		true, // httpOnly - JS cannot read this
	)
	resp := &response.AuthResponse{
		ID:           user.ID,
		Name:         user.Name,
		AccessToken:  accessTokenStr,
		RefreshToken: "ហែងអត់មានការងារធ្វេីហី!!🖕",
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
			"add.payroll", "add.backup", "view.backup", "view.download.backup", "delete.backup", "add.company", "edit.company", "edit.user",
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
	hashedRefresh := utils.HashToken(refreshTokenStr)
	if err := s.db.Where("user_id = ?", user.ID).Delete(&model.Session{}).Error; err != nil {
		log.Printf(err.Error())
		return nil, fmt.Errorf("failed to delete session")
	}

	refreshExpiry := time.Now().Add(time.Duration(refreshtoken) * 24 * time.Hour)

	session := model.Session{
		UserID:       uint(user.ID),
		RefreshToken: string(hashedRefresh),
		TokenPrefix:  tokenPrefix,
		ExpiresAt:    refreshExpiry,
	}
	if err := s.db.Create(&session).Error; err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	maxAge := int(time.Until(refreshExpiry).Seconds())
	c.SetSameSite(http.SameSiteStrictMode)
	c.SetCookie(
		"refresh_token",
		refreshTokenStr,
		maxAge,
		"/",
		"",   // domain: leave empty for current host, or set "yourdomain.com" explicitly
		true, // Secure - HTTPS only
		true, // httpOnly - JS cannot read this
	)
	resp := &response.AuthResponse{
		ID:          user.ID,
		Name:        user.Name,
		AccessToken: accessTokenStr,
		//		RefreshToken: refreshTokenStr,
		Permissions: permissions,
	}

	return resp, nil
}

func (s *authservice) RefreshToken(refreshToken string, c *gin.Context) (*response.AuthResponse, error) {
	if len(refreshToken) < 16 {
		return nil, errors.New("Invalid refresh token")
	}
	prefix := refreshToken[:16]

	var session model.Session
	err := s.db.Where("token_prefix = ?", prefix).First(&session).Error
	if err != nil {
		return nil, errors.New("Invalid or expired refresh token")
	}

	if session.ExpiresAt.Before(time.Now()) {
		return nil, errors.New("Invalid or expired refresh token")
	}

	if !utils.VerifyToken(session.RefreshToken, refreshToken) {
		return nil, errors.New("Invalid or expired refresh token")
	}

	var settings []model.Setting
	if err := s.db.Where("`key` IN ?", []string{
		"ACCESS_TOKEN_EXPIRE_HOURS",
		"REFRESH_TOKEN_EXPIRE_DAYS",
	}).Find(&settings).Error; err != nil {
		return nil, err
	}

	settingMap := make(map[string]string)
	for _, s := range settings {
		settingMap[s.Key] = s.Value
	}

	accesstoken, err := strconv.Atoi(settingMap["ACCESS_TOKEN_EXPIRE_HOURS"])
	if err != nil {
		return nil, err
	}

	refreshtoken, err := strconv.Atoi(settingMap["REFRESH_TOKEN_EXPIRE_DAYS"])
	if err != nil {
		return nil, errors.New("Bad request")
	}

	accessExpiry := time.Now().Add(time.Duration(accesstoken) * time.Minute)
	refreshExpiry := time.Now().Add(time.Duration(refreshtoken) * 24 * time.Hour)
	newRefreshBytes := make([]byte, 32)
	if _, err := rand.Read(newRefreshBytes); err != nil {
		return nil, errors.New("failed to generate refresh token")
	}
	newRefreshStr := hex.EncodeToString(newRefreshBytes)
	newHash := utils.HashToken(newRefreshStr)
	newPrefix := newRefreshStr[:16]

	if err := s.db.Model(&session).Updates(model.Session{
		RefreshToken: newHash,
		TokenPrefix:  newPrefix,
		ExpiresAt:    refreshExpiry,
	}).Error; err != nil {
		return nil, err
	}

	var user model.User
	if err := s.db.Select("id,role_id").Where("id = ?", session.UserID).First(&user).Error; err != nil {
		return nil, err
	}
	claims := jwt.MapClaims{
		"user_id": user.ID,
		"role_id": user.RoleID,
		"exp":     accessExpiry.Unix(),
	}
	accessToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(utils.Jwtkey)
	if err != nil {
		return nil, err
	}

	var permissions []model.Permission
	if err := s.db.Table("permission p").
		Select("p.id AS id, p.name AS name").
		Joins("JOIN role_permission rhp ON rhp.permission_id = p.id").
		Where("rhp.role_id = ? AND p.name IN ?", user.RoleID, []string{
			"add.payroll", "add.backup", "view.backup", "view.download.backup", "delete.backup", "add.company", "edit.company", "edit.user",
		}).
		Scan(&permissions).Error; err != nil {
		return nil, err
	}

	maxAge := int(time.Until(refreshExpiry).Seconds())
	c.SetSameSite(http.SameSiteStrictMode)
	c.SetCookie(
		"refresh_token",
		newRefreshStr,
		maxAge,
		"/",
		"",   // domain: leave empty for current host, or set "yourdomain.com" explicitly
		true, // Secure - HTTPS only
		true, // httpOnly - JS cannot read this
	)

	return &response.AuthResponse{
		AccessToken: accessToken,
		//		RefreshToken: newRefreshStr,
		Permissions: permissions,
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
	basesalaryencrypted, err := helper.EncryptSalary(input.BaseSalary)
	if err != nil {
		return err
	}
	if err != nil {
		return err
	}
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
		BaseSalary:     basesalaryencrypted,
		CompanyID:      companyID,
		QrToken:        qrTokenHash,
		IsVerify:       false,
		QrTokenEncript: qrTokenEncript,
		ManageCompany:  input.ManageCompany,
	}
	if err := tx.Create(&user).Error; err != nil {
		return err
	}

	if input.ManageCompany == 2 && len(input.CompanyIDs) != 0 {
		for i := range input.CompanyIDs {
			usercompany := model.UserCompany{
				UserID:    user.ID,
				CompanyID: *input.CompanyIDs[i],
			}
			if err := tx.Create(&usercompany).Error; err != nil {
				return err
			}
		}
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
		switch user.ManageCompany {
		case 1:
			return query.Where("u.company_id =?", user.CompanyID)
		case 2:
			var companyIDs []int
			db.Model(&model.UserCompany{}).Where("user_id =?", user.ID).Pluck("company_id", &companyIDs)
			if len(companyIDs) == 0 {
				return query.Where("1 = 0")
			}
			return query.Where("u.company_id IN ?", companyIDs)
		}
		return query
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
			c.currency AS currency,
			u.manage_company AS manage_company
        `).
		Joins("LEFT JOIN role r ON r.id = u.role_id").
		Joins("LEFT JOIN company c ON c.id = u.company_id")

	userquery = applyAccessFilter(userquery, s.db, user.Role, user)
	userquery = applyCommonFilter(userquery, filter)
	userquery = userquery.Order("id DESC")
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
			log.Printf(err.Error())
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

	var usercompany []model.UserCompany
	usercompanyquery := s.db.WithContext(ctx).Table("user_company uc").Select(`
		uc.user_id AS user_id,
		uc.company_id AS company_id
	`).Where("uc.user_id IN ?", userIDs)

	if err := usercompanyquery.Scan(&usercompany).Error; err != nil {
		return nil, nil, err
	}
	usercompanybyuserID := make(map[int][]model.UserCompany, len(users))
	for _, r := range usercompany {
		usercompanybyuserID[r.UserID] = append(usercompanybyuserID[r.UserID], r)

	}

	for i, a := range users {
		var ids []int
		for _, uc := range usercompanybyuserID[a.ID] {
			ids = append(ids, uc.CompanyID)
		}
		users[i].CompanyIDs = ids
	}

	return users, helper.BuildPaginationMeta(pf, totalCount), nil
}

var ErrCannotToggleOwnStatus = errors.New("cannot toggle your own status")

func (s *authservice) ToggleUserStatus(ctx context.Context, id int, userID int) error {
	if id == userID {
		return ErrCannotToggleOwnStatus
	}
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
	}
	if input.RoleID != nil {
		updates["role_id"] = *input.RoleID
	}
	if input.CompanyID != nil {
		updates["company_id"] = *input.CompanyID
	}
	if input.Name != nil {
		updates["name"] = *input.Name
	}
	if input.Gender != nil {
		updates["gender"] = *input.Gender
	}
	if input.ManageCompany != nil {
		updates["manage_company"] = *input.ManageCompany
	}
	if input.BaseSalary != nil {
		encrypted, err := helper.EncryptSalary(*input.BaseSalary)
		if err != nil {
			return err
		}
		updates["base_salary"] = encrypted
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

	if len(updates) > 0 {
		if err := tx.Model(&model.User{}).Where("id = ?", id).Updates(updates).Error; err != nil {
			slog.Error("failed to update user", "error", err)
			return err
		}
	}

	if *input.ManageCompany != 2 {
		if err := tx.Where("user_id = ?", id).Delete(&model.UserCompany{}).Error; err != nil {
			return err
		}
	} else if *input.ManageCompany == 2 {
		for _, cid := range input.CompanyIDs {
			if cid == nil {
				return errors.New("company_ids must not contain null values")
			}
			uc := model.UserCompany{
				UserID:    id,
				CompanyID: *cid,
			}
			if err := tx.Create(&uc).Error; err != nil {
				return err
			}
		}
	}

	if err := tx.Commit().Error; err != nil {
		return err
	}
	committed = true
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

func (s *authservice) DeleteUser(ctx context.Context, id int, userIDlogin int) error {
	var target model.User
	if err := s.db.Preload("Role").First(&target, id).Error; err != nil {
		return err
	}

	var usrlog model.User
	if err := s.db.Preload("Role").Find(&usrlog, userIDlogin).Error; err != nil {
		return err
	}

	if err := helper.CanManageUser(usrlog.Role, target.Role); err != nil {
		return err
	}

	return s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where(
			"attendance_id IN (SELECT id FROM attendance WHERE user_id = ?)", id,
		).Delete(&model.AttendanceRecord{}).Error; err != nil {
			return err
		}

		if err := tx.Where("user_id = ?", id).Delete(&model.Attendance{}).Error; err != nil {
			return err
		}

		if err := tx.Where("user_id = ?", id).Delete(&model.Payroll{}).Error; err != nil {
			return err
		}

		if err := tx.Where("user_id = ?", id).Delete(&model.Session{}).Error; err != nil {
			return err
		}

		if err := tx.Where("user_id = ?", id).Delete(&model.Shift{}).Error; err != nil {
			return err
		}

		if err := tx.Where("id = ?", id).Delete(&model.User{}).Error; err != nil {
			return err
		}

		return nil
	})
}
