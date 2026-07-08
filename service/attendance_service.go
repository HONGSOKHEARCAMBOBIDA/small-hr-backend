package service

import (
	"context"
	"errors"
	"fmt"
	"html"
	"mysql/config"
	"mysql/helper"
	"mysql/model"
	"mysql/request"
	"mysql/response"
	"mysql/utils"
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type AttendanceService interface {
	CreateAttendance(ctx context.Context, id int, input request.AttendanceRequestCreate) error
	GetAttendance(ctx context.Context, id int, pf request.Pagination, filter map[string]string) ([]response.AttendanceResponse, *model.PaginationMetadata, error)
	GetAttendanceDraft(ctx context.Context, id int) (response.AttendanceResponseDraft, error)
	GetAttendancePDF(ctx context.Context, id int, pf request.Pagination, filter map[string]string) ([]response.AttendanceResponseGenerate, *model.PaginationMetadata, error)
	DeleteAttendance(ctx context.Context, id int) error
}

type attendanceservice struct {
	db *gorm.DB
}

func NewAttendanceService() AttendanceService {
	return &attendanceservice{
		db: config.DB,
	}
}

const (
	MorningShiftOnly = 2
	EveningShiftOnly = 3
	FullShift        = 1
)

const (
	RecordMorningCheckIn  = 1
	RecordMorningCheckOut = 2
	RecordEveningCheckIn  = 3
	RecordEveningCheckOut = 4
)

type LeaveSession int

const (
	LeaveNone LeaveSession = iota
	LeaveFull
	LeaveMorning
	LeaveEvening
)

var recordTypeLabel = map[int]string{
	RecordMorningCheckIn:  "ចូលធ្វេីការវែនទី១",
	RecordMorningCheckOut: "ចេញពីធ្វេីការវែនទី១",
	RecordEveningCheckIn:  "ចូលធ្វេីការវែនទី២",
	RecordEveningCheckOut: "ចេញពីធ្វេីការវែនទី២",
}

var checkTypeLabel = map[int]string{
	RecordMorningCheckIn:  "ចូលធ្វេីការវែនទី១",
	RecordMorningCheckOut: "ចេញពីធ្វើការវែនទី១",
	RecordEveningCheckIn:  "ចូលធ្វេីការវែនទី២",
	RecordEveningCheckOut: "ចេញពីធ្វេីការវែនទី២",
}

type sessionConfig struct {
	scheduledTime string
	isCheckIn     bool
	recordType    int
}

func buildSessionV2(shift model.Shift, leave LeaveSession) ([]sessionConfig, error) {
	if leave == LeaveFull {
		return nil, errors.New("today is a full-day approved leave")
	}
	switch shift.ShiftType {
	case MorningShiftOnly:
		if shift.CheckIn1 == nil || shift.CheckOut1 == nil {
			return nil, errors.New("shift type 2: CheckIn1 or CheckOut1 is not configured")
		}
		return []sessionConfig{
			{scheduledTime: *shift.CheckIn1, isCheckIn: true, recordType: RecordMorningCheckIn},
			{scheduledTime: *shift.CheckOut1, isCheckIn: false, recordType: RecordMorningCheckOut},
		}, nil
	case EveningShiftOnly:
		if shift.CheckIn2 == nil || shift.CheckOut2 == nil {
			return nil, errors.New("shift type 3: CheckIn2 or CheckOut2 is not configured")
		}
		return []sessionConfig{
			{scheduledTime: *shift.CheckIn2, isCheckIn: true, recordType: RecordEveningCheckIn},
			{scheduledTime: *shift.CheckOut2, isCheckIn: false, recordType: RecordEveningCheckOut},
		}, nil
	case FullShift:
		switch leave {
		case LeaveMorning:
			if shift.CheckIn2 == nil || shift.CheckOut2 == nil {
				return nil, errors.New("shift: CheckIn2 or CheckOut2 is not configured")
			}
			return []sessionConfig{
				{scheduledTime: *shift.CheckIn2, isCheckIn: true, recordType: RecordEveningCheckIn},
				{scheduledTime: *shift.CheckOut2, isCheckIn: false, recordType: RecordEveningCheckOut},
			}, nil
		case LeaveEvening:
			if shift.CheckIn1 == nil || shift.CheckOut1 == nil {
				return nil, errors.New("shift: CheckIn1 or CheckOut1 is not configure")
			}
			return []sessionConfig{
				{scheduledTime: *shift.CheckIn1, isCheckIn: true, recordType: RecordMorningCheckIn},
				{scheduledTime: *shift.CheckOut1, isCheckIn: true, recordType: RecordEveningCheckOut},
			}, nil
		default:
			if shift.CheckIn1 == nil || shift.CheckOut1 == nil || shift.CheckIn2 == nil || shift.CheckOut2 == nil {
				return nil, errors.New("shift: one or more check-in/out times are not configured")
			}
			return []sessionConfig{
				{scheduledTime: *shift.CheckIn1, isCheckIn: true, recordType: RecordMorningCheckIn},
				{scheduledTime: *shift.CheckOut1, isCheckIn: false, recordType: RecordMorningCheckOut},
				{scheduledTime: *shift.CheckIn2, isCheckIn: true, recordType: RecordEveningCheckIn},
				{scheduledTime: *shift.CheckOut2, isCheckIn: false, recordType: RecordEveningCheckOut},
			}, nil
		}
	default:
		return nil, fmt.Errorf("unsupported shift type: %d", shift.ShiftType)
	}
}

func buildSession(shift model.Shift) ([]sessionConfig, error) {
	switch shift.ShiftType {
	case MorningShiftOnly:
		if shift.CheckIn1 == nil || shift.CheckOut1 == nil {
			return nil, errors.New("shift type 2: CheckIn1 or CheckOut1 is not configured")
		}
		return []sessionConfig{
			{scheduledTime: *shift.CheckIn1, isCheckIn: true, recordType: RecordMorningCheckIn},
			{scheduledTime: *shift.CheckOut1, isCheckIn: false, recordType: RecordMorningCheckOut},
		}, nil
	case EveningShiftOnly:
		if shift.CheckIn2 == nil || shift.CheckOut2 == nil {
			return nil, errors.New("shift type 3: CheckIn2 or CheckOut2 is not configured")
		}
		return []sessionConfig{
			{scheduledTime: *shift.CheckIn2, isCheckIn: true, recordType: RecordEveningCheckIn},
			{scheduledTime: *shift.CheckOut2, isCheckIn: false, recordType: RecordEveningCheckOut},
		}, nil
	case FullShift:
		if shift.CheckIn1 == nil || shift.CheckOut1 == nil || shift.CheckIn2 == nil || shift.CheckOut2 == nil {
			return nil, errors.New("shift: one or more check-in/out times are not configured")
		}
		return []sessionConfig{
			{scheduledTime: *shift.CheckIn1, isCheckIn: true, recordType: RecordMorningCheckIn},
			{scheduledTime: *shift.CheckOut1, isCheckIn: false, recordType: RecordMorningCheckOut},
			{scheduledTime: *shift.CheckIn2, isCheckIn: true, recordType: RecordEveningCheckIn},
			{scheduledTime: *shift.CheckOut2, isCheckIn: false, recordType: RecordEveningCheckOut},
		}, nil
	default:
		return nil, fmt.Errorf("unsupported shift type: %d", shift.ShiftType)
	}
}

func (s *attendanceservice) getApprovedLeaveSession(ctx context.Context, userID int) (LeaveSession, error) {
	currentDate := time.Now().Format("2006-01-02")

	var leaveRequest model.LeaveRequest
	err := s.db.WithContext(ctx).
		Preload("LeaveDeductType").
		Where("user_id = ? AND status = ? AND start_date <= ? AND end_date >= ?",
			userID, 2, currentDate, currentDate).
		First(&leaveRequest).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// no approved leave today -> normal full-day attendance flow
			return LeaveNone, nil
		}
		return LeaveNone, fmt.Errorf("failed to load leave request: %w", err)
	}

	deductCode := leaveRequest.LeaveDeductType.Code
	switch deductCode {
	case "FULL_DAY":
		return LeaveFull, nil
	case "HALF_AM":
		return LeaveMorning, nil
	case "HALF_PM":
		return LeaveEvening, nil
	default:
		return LeaveNone, fmt.Errorf("unknown leave deduct type code: %s", deductCode)
	}
}

func (s *attendanceservice) CreateAttendance(ctx context.Context, id int, input request.AttendanceRequestCreate) error {
	now := time.Now()
	currentDate := now.Format("2006-01-02")
	currentTime := now.Format("15:04:05")
	dayofweek := helper.GetCurrentDay()
	var leaverequest model.LeaveRequest

	if input.IsPermission {
		err := s.db.WithContext(ctx).Where("user_id = ? AND status = ? AND start_date <= ? AND end_date >= ?",
			id, 2, currentDate, currentDate).First(&leaverequest).Error
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errors.New("គ្មានច្បាប់ឬច្បាប់របស់អ្នកមិនទាន់អនុម័ត")
			}
			return fmt.Errorf("failed to load leave request: %w", err)
		}
	}

	var user model.User
	if err := s.db.WithContext(ctx).Preload("Company").First(&user, id).Error; err != nil {
		return fmt.Errorf("failed to load user: %w", err)
	}

	var shift model.Shift
	if err := s.db.WithContext(ctx).Where("user_id = ? AND day = ?", user.ID, dayofweek).First(&shift).Error; err != nil {
		return fmt.Errorf("failed to load shift: %w", err)
	}
	if shift.IsDayoff {
		return errors.New("today is dayoff")
	}

	leave, err := s.getApprovedLeaveSession(ctx, user.ID)
	if err != nil {
		return err
	}

	sessions, err := buildSessionV2(shift, leave)
	if err != nil {
		return err
	}

	companyLat, err := strconv.ParseFloat(user.Company.Latitude, 64)
	if err != nil {
		return fmt.Errorf("invalid company latitude: %w", err)
	}
	companyLng, err := strconv.ParseFloat(user.Company.Longitude, 64)
	if err != nil {
		return fmt.Errorf("invalid company longitude: %w", err)
	}
	userLat, err := strconv.ParseFloat(input.Latitude, 64)
	if err != nil {
		return fmt.Errorf("invalid user latitude: %w", err)
	}
	userLng, err := strconv.ParseFloat(input.Longitude, 64)
	if err != nil {
		return fmt.Errorf("invalid user longitude :%w", err)
	}
	radius, err := strconv.ParseFloat(user.Company.Radius, 64)
	if err != nil {
		return fmt.Errorf("invalid company redius: %w", err)
	}
	distance := utils.CalculateDistance(companyLat, companyLng, userLat, userLng)
	inzone := distance <= radius

	if !inzone && user.Company.CanScanOutsize == 0 {
		return errors.New("អ្នកមិនអាចស្កែនក្រៅតំបន់ក្រុមហ៊ុនបានទេ")
	}

	var current sessionConfig
	var record model.AttendanceRecord
	// var attendanceID uint
	// var justCompleted bool

	txErr := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var attendance model.Attendance
		err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("user_id = ? AND check_date = ?", user.ID, currentDate).First(&attendance).Error
		switch {
		case err == nil:

		case errors.Is(err, gorm.ErrRecordNotFound):
			attendance = model.Attendance{
				UserID:    user.ID,
				CheckDate: currentDate,
				Status:    "WORKING",
			}
			if err := tx.Create(&attendance).Error; err != nil {
				return fmt.Errorf("failed to create attendance: %w", err)
			}
		default:
			return fmt.Errorf("failed to load attendance: %w", err)
		}
		var existingRecords []model.AttendanceRecord
		if err := tx.Where("attendance_id = ?", attendance.ID).Order("id ASC").Find(&existingRecords).Error; err != nil {
			return fmt.Errorf("failed to load attendance :%w", err)
		}
		recordCount := len(existingRecords)
		maxRecords := len(sessions)

		if recordCount >= maxRecords {
			if err := tx.Model(&model.Attendance{}).Where("id = ?", attendance.ID).
				Update("status", "COMPLETE").Error; err != nil {
				return fmt.Errorf("failed to update attendance status:%w", err)
			}
			return errors.New("all check-in and check-out completed for today")
		}
		current = sessions[recordCount]
		attendanceType := helper.DetermineAttendanceType(currentTime, current.scheduledTime, current.isCheckIn)
		record = model.AttendanceRecord{
			AttendanceID:   attendance.ID,
			ShiftID:        shift.ID,
			AttendanceType: attendanceType,
			Reason:         input.Reason,
			CheckTime:      currentTime,
			Type:           current.recordType,
			Inzone:         inzone,
			Latitude:       input.Latitude,
			Longitude:      input.Longitude,
			IsPermission:   input.IsPermission,
		}
		if err := tx.Create(&record).Error; err != nil {
			return fmt.Errorf("failed to created attendance record: %w", err)
		}
		// attendanceID = uint(attendance.ID)
		if recordCount+1 >= maxRecords {
			if err := tx.Model(&model.Attendance{}).Where("id = ?", attendance.ID).
				Update("status", "COMPLETE").Error; err != nil {
				return fmt.Errorf("faild to update attendance status %w", err)
			}
			// justCompleted = true
		}
		return nil
	})
	if txErr != nil {
		return txErr
	}
	s.notifyTelegram(user, shift, current, record, currentTime, inzone, input.Reason)

	// _ = attendanceID
	// _ = justCompleted
	return nil
}

func (s *attendanceservice) notifyTelegram(
	user model.User,
	shift model.Shift,
	current sessionConfig,
	record model.AttendanceRecord,
	currentTime string,
	inZone bool,
	reason string,
) {
	if user.Company.GroupChatID == nil || user.Company.BotToken == nil {
		return
	}

	groupChatID, err := utils.DecryptChatID(*user.Company.GroupChatID)
	if err != nil {
		fmt.Errorf("failed to decrypt group chat id %w", err)
		return
	}
	botToken, err := utils.DecryptBotToken(*user.Company.BotToken)
	if err != nil {
		fmt.Errorf("failed to decrypt bot token %w", err)
	}

	checktype, ok := checkTypeLabel[current.recordType]
	if !ok {
		fmt.Errorf("unknown record type for notification %w", err)
		return
	}

	lateText := attendanceTypeLabel(record.AttendanceType)

	zoneText := "📍 ស្កែនក្នុងតំបន់ក្រុមហ៊ុន"
	if !inZone {
		zoneText = "⚠️ ស្កែនក្រៅតំបន់ក្រុមហ៊ុន"
	}

	// Escape user-controlled text before embedding into an HTML-formatted message.
	safeReason := html.EscapeString(reason)
	safeName := html.EscapeString(user.Name)
	safeCompany := html.EscapeString(user.Company.Name)

	message := fmt.Sprintf(
		"<b>%s</b>\n\n"+
			"👤 ឈ្មោះ: %s\n"+
			"🏢 សាខា: %s\n"+
			"🕒 ត្រូវស្កែន: %s\n"+
			"🕒 បានស្កែន: %s\n"+
			"%s\n"+
			"%s\n"+
			"មូលហេតុ: %s\n",
		checktype,
		safeName,
		safeCompany,
		current.scheduledTime,
		currentTime,
		lateText,
		zoneText,
		safeReason,
	)

	go func() {
		defer func() {
			if r := recover(); r != nil {
				fmt.Errorf("panic while sending telegram message")
			}
		}()
		helper.SendTelegramMessage(message, groupChatID, botToken)
	}()
}

func attendanceTypeLabel(attendanceType int) string {
	switch attendanceType {
	case 1:
		return "🟢 ចូលធ្វើការមុនម៉ោង"
	case 2:
		return "🤭 ចូលធ្វើការទាន់ម៉ោង"
	case 3:
		return "🔴 ចូលធ្វើការយឺត"
	case 4:
		return "🤫 ចេញពីធ្វើការមុនម៉ោង"
	case 5:
		return "😴 ចេញពីធ្វើការត្រឹមម៉ោង"
	case 6:
		return "😓 ចេញពីធ្វើការក្រោយម៉ោង"
	default:
		return ""
	}
}

// func (s *attendanceservice) CreateAttendance(ctx context.Context, id int, input request.AttendanceRequestCreate) error {
// 	tx := s.db.WithContext(ctx).Begin()
// 	if tx.Error != nil {
// 		return tx.Error
// 	}

// 	committed := false
// 	defer func() {
// 		if r := recover(); r != nil {
// 			tx.Rollback()
// 		} else if !committed {
// 			tx.Rollback()
// 		}
// 	}()

// 	now := time.Now()
// 	currentDate := now.Format("2006-01-02")
// 	currentTime := now.Format("15:04:05")
// 	dayofweek := helper.GetCurrentDay()

// 	var user model.User
// 	if err := s.db.WithContext(ctx).Preload("Company").First(&user, id).Error; err != nil {
// 		tx.Rollback()
// 		return err
// 	}

// 	var shift model.Shift
// 	if err := tx.Where("user_id = ? AND day = ?", user.ID, dayofweek).First(&shift).Error; err != nil {
// 		tx.Rollback()
// 		return err
// 	}
// 	if shift.IsDayoff {
// 		tx.Rollback()
// 		return errors.New("today is dayoff")
// 	}

// 	companyLat, err := strconv.ParseFloat(user.Company.Latitude, 64)
// 	if err != nil {
// 		tx.Rollback()
// 		return err
// 	}
// 	companyLng, err := strconv.ParseFloat(user.Company.Longitude, 64)
// 	if err != nil {
// 		tx.Rollback()
// 		return err
// 	}
// 	userLat, err := strconv.ParseFloat(input.Latitude, 64)
// 	if err != nil {
// 		tx.Rollback()
// 		return err
// 	}
// 	userLng, err := strconv.ParseFloat(input.Longitude, 64)
// 	if err != nil {
// 		tx.Rollback()
// 		return err
// 	}
// 	radius, err := strconv.ParseFloat(user.Company.Radius, 64)
// 	if err != nil {
// 		tx.Rollback()
// 		return err
// 	}

// 	distance := utils.CalculateDistance(companyLat, companyLng, userLat, userLng)
// 	inZone := distance <= radius

// 	if !inZone && user.Company.CanScanOutsize == 0 {
// 		tx.Rollback()
// 		return errors.New("អ្នកមិនអាចស្កែនក្រៅតំបន់ក្រុមហ៊ុនបានទេ")
// 	}

// 	var attendance model.Attendance
// 	err = tx.Where("user_id = ? AND check_date = ?", user.ID, currentDate).First(&attendance).Error
// 	if err != nil {
// 		if !errors.Is(err, gorm.ErrRecordNotFound) {
// 			tx.Rollback()
// 			return err
// 		}

// 		attendance = model.Attendance{
// 			UserID:    user.ID,
// 			CheckDate: currentDate,
// 			Status:    "WORKING",
// 		}
// 		if err := tx.Create(&attendance).Error; err != nil {
// 			tx.Rollback()
// 			return err
// 		}
// 	}

// 	var existingRecords []model.AttendanceRecord
// 	if err := tx.Where("attendance_id = ?", attendance.ID).
// 		Order("id ASC").
// 		Find(&existingRecords).Error; err != nil {
// 		tx.Rollback()
// 		return err
// 	}

// 	recordCount := len(existingRecords)

// 	// Build session list based on ShiftType
// 	// ShiftType 1 = Full day    → CheckIn1, CheckOut1, CheckIn2, CheckOut2
// 	// ShiftType 2 = Morning only → CheckIn1, CheckOut1
// 	// ShiftType 3 = Evening only → CheckIn2, CheckOut2

// 	type sessionConfig struct {
// 		scheduledTime string
// 		isCheckIn     bool
// 		recordType    int // 1=CheckIn1, 2=CheckOut1, 3=CheckIn2, 4=CheckOut2
// 	}

// 	var sessions []sessionConfig

// 	switch shift.ShiftType {

// 	case 2:
// 		if shift.CheckIn1 == nil || shift.CheckOut1 == nil {
// 			return errors.New("ធ្វេីការវែនព្រឹកតែមិនទាន់ដាក់ម៉ោងចេញចូល")
// 		}
// 		sessions = []sessionConfig{
// 			{scheduledTime: *shift.CheckIn1, isCheckIn: true, recordType: 1},
// 			{scheduledTime: *shift.CheckOut1, isCheckIn: false, recordType: 2},
// 		}
// 	case 3:
// 		if shift.CheckIn2 == nil || shift.CheckOut2 == nil {
// 			return errors.New("ធ្វេីការវែនល្ងាចតែមិនទាន់ដាក់ម៉ោងចេញចូល")
// 		}
// 		sessions = []sessionConfig{
// 			{scheduledTime: *shift.CheckIn2, isCheckIn: true, recordType: 3},
// 			{scheduledTime: *shift.CheckOut2, isCheckIn: false, recordType: 4},
// 		}
// 	default:
// 		if shift.CheckIn1 == nil || shift.CheckOut1 == nil || shift.CheckIn2 == nil || shift.CheckOut2 == nil {
// 			return errors.New("គ្មានម៉ោងធ្វេីការ")
// 		}
// 		sessions = []sessionConfig{
// 			{scheduledTime: *shift.CheckIn1, isCheckIn: true, recordType: 1},
// 			{scheduledTime: *shift.CheckOut1, isCheckIn: false, recordType: 2},
// 			{scheduledTime: *shift.CheckIn2, isCheckIn: true, recordType: 3},
// 			{scheduledTime: *shift.CheckOut2, isCheckIn: false, recordType: 4},
// 		}
// 	}

// 	maxRecords := len(sessions)

// 	if recordCount >= maxRecords {
// 		if err := tx.Model(&model.Attendance{}).
// 			Where("id = ?", attendance.ID).
// 			Update("status", "COMPLETE").Error; err != nil {
// 			return fmt.Errorf("failed to update attendance status: %w", err)
// 		}
// 		if err := tx.Commit().Error; err != nil {
// 			return err
// 		}
// 		committed = true
// 		return errors.New("all check-ins and check-outs completed for today")
// 	}

// 	current := sessions[recordCount]

// 	attendanceType := helper.DetermineAttendanceType(currentTime, current.scheduledTime, current.isCheckIn)

// 	record := model.AttendanceRecord{
// 		AttendanceID:   attendance.ID,
// 		ShiftID:        shift.ID,
// 		AttendanceType: attendanceType,
// 		Reason:         input.Reason,
// 		CheckTime:      currentTime,
// 		Type:           current.recordType,
// 		Inzone:         inZone,
// 		Latitude:       input.Latitude,
// 		Longitude:      input.Longitude,
// 	}
// 	if err := tx.Create(&record).Error; err != nil {
// 		return err
// 	}

// 	if recordCount+1 >= maxRecords {
// 		if err := tx.Model(&model.Attendance{}).
// 			Where("id = ?", attendance.ID).
// 			Update("status", "COMPLETE").Error; err != nil {
// 			return fmt.Errorf("failed to update attendance status: %w", err)
// 		}
// 	}

// 	workTime := fmt.Sprintf("%s", current.scheduledTime)
// 	lateText := ""

// 	switch attendanceType {
// 	case 1:
// 		lateText = "🟢 ចូលធ្វើការមុនម៉ោង"
// 	case 2:
// 		lateText = "🤭 ចូលធ្វើការទាន់ម៉ោង"
// 	case 3:
// 		lateText = "🔴 ចូលធ្វើការយឺត"
// 	case 4:
// 		lateText = "🤫 ចេញពីធ្វើការមុនម៉ោង"
// 	case 5:
// 		lateText = "😴 ចេញពីធ្វើការត្រឹមម៉ោង"
// 	case 6:
// 		lateText = "😓 ចេញពីធ្វើការក្រោយម៉ោង"
// 	}

// 	zoneText := "📍 ស្កែនក្នុងតំបន់ក្រុមហ៊ុន"
// 	if !inZone {
// 		zoneText = "⚠️ ស្កែនក្រៅតំបន់ក្រុមហ៊ុន"
// 	}

// 	checktype := ""
// 	switch current.recordType {
// 	case 1:
// 		checktype = "ចូលធ្វេីការវែនទី១"
// 	case 2:
// 		checktype = "ចេញពីធ្វើការវែនទី១"
// 	case 3:
// 		checktype = "ចូលធ្វេីការវែនទី២"
// 	case 4:
// 		checktype = "ចេញពីធ្វេីការវែនទី២"
// 	default:
// 		return fmt.Errorf("unknown record type: %d", current.recordType)
// 	}
// 	GroupChatIDDecrypt, err := utils.DecryptChatID(*user.Company.GroupChatID)
// 	BotTokenDecrypt, err := utils.DecryptBotToken(*user.Company.BotToken)
// 	message := fmt.Sprintf(
// 		"<b>%s</b>\n\n"+
// 			"👤 ឈ្មោះ: %s\n"+
// 			"🏢 សាខា: %s\n"+
// 			"🕒 ម៉ោងត្រូវស្កែន: %s\n"+
// 			"🕒 ម៉ោងបានស្កែន: %s\n"+
// 			"%s\n"+
// 			"%s\n"+
// 			"មូលហេតុ: %s\n",
// 		checktype,
// 		user.Name,
// 		user.Company.Name,
// 		workTime,
// 		now.Format("15:04:05"),
// 		lateText,
// 		zoneText,
// 		input.Reason,
// 	)
// 	go helper.SendTelegramMessage(message, GroupChatIDDecrypt, BotTokenDecrypt)

// 	if err := tx.Commit().Error; err != nil {
// 		return err
// 	}
// 	committed = true
// 	return nil

// }

func (s *attendanceservice) GetAttendanceDraft(ctx context.Context, id int) (response.AttendanceResponseDraft, error) {
	now := time.Now()
	currentDate := now.Format("2006-01-02")
	dayOfWeek := helper.GetCurrentDay()

	var user model.User
	if err := s.db.WithContext(ctx).Select("id").First(&user, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return response.AttendanceResponseDraft{}, fmt.Errorf("user not found: %w", err)
		}
		return response.AttendanceResponseDraft{}, fmt.Errorf("failed to load user: %w", err)
	}

	var shift model.Shift
	if err := s.db.WithContext(ctx).Where("user_id = ? AND day = ?", user.ID, dayOfWeek).First(&shift).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return response.AttendanceResponseDraft{}, fmt.Errorf("shift not found: %w", err)
		}
		return response.AttendanceResponseDraft{}, fmt.Errorf("failed to load shift: %w", err)
	}

	if shift.IsDayoff {
		return response.AttendanceResponseDraft{}, errors.New("today is a day off")
	}

	leave, err := s.getApprovedLeaveSession(ctx, user.ID)
	if err != nil {
		return response.AttendanceResponseDraft{}, err
	}

	sessions, err := buildSessionV2(shift, leave)
	if err != nil {
		return response.AttendanceResponseDraft{}, err
	}

	var current sessionConfig
	txErr := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var attendance model.Attendance
		var existingRecords []model.AttendanceRecord

		err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("user_id = ? AND check_date = ?", user.ID, currentDate).First(&attendance).Error

		switch {
		case err == nil:
			if err := tx.Where("attendance_id = ?", attendance.ID).Order("id ASC").Find(&existingRecords).Error; err != nil {
				return fmt.Errorf("failed to load attendance records: %w", err)
			}
		case errors.Is(err, gorm.ErrRecordNotFound):
		default:
			return fmt.Errorf("failed to load attendance: %w", err)
		}

		recordCount := len(existingRecords)
		if recordCount >= len(sessions) {
			return errors.New("all attendance sessions for today have already been recorded")
		}
		current = sessions[recordCount]
		return nil
	})
	if txErr != nil {
		return response.AttendanceResponseDraft{}, txErr
	}
	label, ok := recordTypeLabel[current.recordType]
	if !ok {
		return response.AttendanceResponseDraft{}, fmt.Errorf("unknown record type: %d", current.recordType)
	}
	return response.AttendanceResponseDraft{
		Type:          current.recordType,
		TypeString:    label,
		ScheduledTime: current.scheduledTime,
	}, nil

}

// func (s *attendanceservice) GetAttendanceDraft(ctx context.Context, id int) (response.AttendanceResponseDraft, error) {
// 	now := time.Now()
// 	currentDate := now.Format("2006-01-02")
// 	dayofweek := helper.GetCurrentDay()

// 	var user model.User
// 	if err := s.db.WithContext(ctx).Select("id").First(&user, id).Error; err != nil {
// 		return response.AttendanceResponseDraft{}, fmt.Errorf("user not found: %w", err)
// 	}

// 	var shift model.Shift
// 	if err := s.db.WithContext(ctx).Where("user_id = ? AND day = ?", user.ID, dayofweek).First(&shift).Error; err != nil {
// 		return response.AttendanceResponseDraft{}, fmt.Errorf("shift not found: %w", err)
// 	}

// 	if shift.IsDayoff {
// 		return response.AttendanceResponseDraft{}, errors.New("today is a day off")
// 	}

// 	var attendance model.Attendance
// 	var existingRecords []model.AttendanceRecord

// 	err := s.db.WithContext(ctx).Where("user_id = ? AND check_date = ?", user.ID, currentDate).First(&attendance).Error
// 	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
// 		return response.AttendanceResponseDraft{}, fmt.Errorf("failed to load attendance: %w", err)
// 	}
// 	if err == nil {
// 		if err := s.db.WithContext(ctx).Where("attendance_id = ?", attendance.ID).
// 			Order("id ASC").
// 			Find(&existingRecords).Error; err != nil {
// 			return response.AttendanceResponseDraft{}, fmt.Errorf("failed to load attendance records: %w", err)
// 		}
// 	}

// 	type sessionConfig struct {
// 		scheduledTime string
// 		isCheckIn     bool
// 		recordType    int
// 	}

// 	var sessions []sessionConfig

// 	switch shift.ShiftType {
// 	case MorningShiftOnly:
// 		if shift.CheckIn1 == nil || shift.CheckOut1 == nil {
// 			return response.AttendanceResponseDraft{}, errors.New("shift type 2: CheckIn1 or CheckOut1 is not configured")
// 		}
// 		sessions = []sessionConfig{
// 			{scheduledTime: *shift.CheckIn1, isCheckIn: true, recordType: 1},
// 			{scheduledTime: *shift.CheckOut1, isCheckIn: false, recordType: 2},
// 		}
// 	case EveningShiftOnly:
// 		if shift.CheckIn2 == nil || shift.CheckOut2 == nil {
// 			return response.AttendanceResponseDraft{}, errors.New("shift type 3: CheckIn2 or CheckOut2 is not configured")
// 		}
// 		sessions = []sessionConfig{
// 			{scheduledTime: *shift.CheckIn2, isCheckIn: true, recordType: 3},
// 			{scheduledTime: *shift.CheckOut2, isCheckIn: false, recordType: 4},
// 		}
// 	case FullShift:
// 		if shift.CheckIn1 == nil || shift.CheckOut1 == nil || shift.CheckIn2 == nil || shift.CheckOut2 == nil {
// 			return response.AttendanceResponseDraft{}, errors.New("shift: one or more check-in/out times are not configured")
// 		}
// 		sessions = []sessionConfig{
// 			{scheduledTime: *shift.CheckIn1, isCheckIn: true, recordType: 1},
// 			{scheduledTime: *shift.CheckOut1, isCheckIn: false, recordType: 2},
// 			{scheduledTime: *shift.CheckIn2, isCheckIn: true, recordType: 3},
// 			{scheduledTime: *shift.CheckOut2, isCheckIn: false, recordType: 4},
// 		}
// 	default:

// 	}

// 	recordCount := len(existingRecords)
// 	if recordCount >= len(sessions) {
// 		return response.AttendanceResponseDraft{}, errors.New("all attendance sessions for today have already been recorded")
// 	}

// 	current := sessions[recordCount]

// 	var checktype string
// 	switch current.recordType {
// 	case 1:
// 		checktype = "ចូលធ្វេីការវែនទី១"
// 	case 2:
// 		checktype = "ចេញពីធ្វេីការវែនទី១"
// 	case 3:
// 		checktype = "ចូលធ្វេីការវែនទី២"
// 	case 4:
// 		checktype = "ចេញពីធ្វេីការវែនទី២"
// 	default:
// 		return response.AttendanceResponseDraft{}, fmt.Errorf("unknown record type: %d", current.recordType)
// 	}

// 	return response.AttendanceResponseDraft{
// 		Type:          current.recordType,
// 		TypeString:    checktype,
// 		ScheduledTime: current.scheduledTime,
// 	}, nil
// }

func applyAccessFilterAttendance(query *gorm.DB, db *gorm.DB, role model.Role, user model.User) *gorm.DB {
	if role.Level > RoleLevelStaft && role.Level <= RoleLevelDeveloper {
		switch user.ManageCompany {
		case ManageOneCompany:
			return query.Where("u.company_id =?", user.CompanyID)
		case ManageMultipleCompany:
			var companyIDs []int
			db.Model(&model.UserCompany{}).Where("user_id =?", user.ID).Pluck("company_id", &companyIDs)
			if len(companyIDs) == 0 {
				return query.Where("1 = 0")
			}
			return query.Where("u.company_id IN ?", companyIDs)
		case ManageAllCompany:
			return query
		default:
			return query.Where("1 = 0")
		}
	} else if role.Level <= RoleLevelStaft {
		return query.Where("u.id =?", user.ID)
	} else if role.Level > RoleLevelManager {
		return query
	}

	return query
}

func applyCommonFilterAttendance(query *gorm.DB, filter map[string]string) *gorm.DB {
	for key, value := range filter {
		value = strings.TrimSpace(value)
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
		case "check_date":
			query = query.Where("a.check_date >=?", value)
		}
	}
	return query
}

func (s *attendanceservice) GetAttendance(ctx context.Context, id int, pf request.Pagination, filter map[string]string) ([]response.AttendanceResponse, *model.PaginationMetadata, error) {
	var attendance []response.AttendanceResponse
	var user model.User
	if err := s.db.WithContext(ctx).Preload("Role").First(&user, id).Error; err != nil {
		return nil, nil, err
	}
	offset := (pf.Page - 1) * pf.PageSize
	attendancequery := s.db.WithContext(ctx).Table("attendance a").
		Select(`
		a.id AS id,
		u.id AS user_id,
		u.name AS name,
		u.gender AS gender,
		c.id AS company_id,
		c.name AS company_name,
		r.id AS role_id,
		r.display_name AS role_name,
		a.check_date AS check_date,
		a.status AS status
	`).
		Joins("LEFT JOIN user u ON u.id = a.user_id").
		Joins("LEFT JOIN company c ON c.id = u.company_id").
		Joins("LEFT JOIN role r ON r.id = u.role_id")

	attendancequery = attendancequery.Order("a.id DESC")

	attendancequery = applyAccessFilterAttendance(attendancequery, s.db, user.Role, user)
	attendancequery = applyCommonFilterAttendance(attendancequery, filter)

	var totalCount int64
	countQuery := attendancequery.Session(&gorm.Session{})
	if err := countQuery.Count(&totalCount).Error; err != nil {
		return nil, nil, err
	}
	if err := attendancequery.Offset(offset).Limit(pf.PageSize).Scan(&attendance).Error; err != nil {
		return nil, nil, err
	}

	for i := range attendance {
		attendance[i].GenderString = helper.Gender(attendance[i].Gender)
		attendance[i].CheckDate = helper.FormatDate(attendance[i].CheckDate)
	}

	if len(attendance) == 0 {
		return attendance, helper.BuildPaginationMeta(pf, totalCount), nil
	}

	attendanceIDs := make([]int, len(attendance))
	for i, a := range attendance {
		attendanceIDs[i] = a.ID
	}

	var attendancerecords []response.AttendanceRecordResponse

	attendancerecordquery := s.db.WithContext(ctx).Table("attendance_record ar").
		Select(`
        ar.id AS id,
        ar.attendance_id AS attendance_id,
        s.day AS day,
        s.check_in1 AS check_in1,
        s.check_out1 AS check_out1,
        s.check_in2 AS check_in2,
        s.check_out2 AS check_out2,
        at.id AS attendance_type,
        at.name AS attendance_type_name,
        ar.resean AS reason,
        ar.check_time AS check_time,
        ar.type AS type,
        ar.inzone AS inzone,
		ar.latitdude AS latitdude,
		ar.longitude AS longitude
    `).
		Joins("LEFT JOIN shift s ON s.id = ar.shift_id").
		Joins("LEFT JOIN attendance_type at ON at.id = ar.attendance_type").
		Where("ar.attendance_id IN ?", attendanceIDs)

	if err := attendancerecordquery.Scan(&attendancerecords).Error; err != nil {
		return nil, nil, err
	}

	for i := range attendancerecords {
		r := &attendancerecords[i]
		r.DayString = helper.DayKhmer(r.Day)
		r.TypeString = helper.TypeFormat(r.Type)
		r.ScheduledTime = helper.DetermineScheduledTime(
			r.Type,
			r.CheckIn1,
			r.CheckOut1,
			r.CheckIn2,
			r.CheckOut2,
		)
		isCheckIn := r.Type == 1 || r.Type == 3
		r.TimeDiff = helper.CalcTimeDiff(r.CheckTime, r.ScheduledTime, isCheckIn)
	}

	recordByAttendanceID := make(map[int][]response.AttendanceRecordResponse, len(attendance))
	for _, r := range attendancerecords {
		recordByAttendanceID[r.AttendanceID] = append(recordByAttendanceID[r.AttendanceID], r)
	}

	for i, a := range attendance {
		attendance[i].AttendanceRecordResponse = recordByAttendanceID[a.ID]
	}

	return attendance, helper.BuildPaginationMeta(pf, totalCount), nil
}

func (s *attendanceservice) GetAttendancePDF(ctx context.Context, id int, pf request.Pagination, filter map[string]string) ([]response.AttendanceResponseGenerate, *model.PaginationMetadata, error) {
	var attendance []response.AttendanceResponseGenerate
	var user model.User
	if err := s.db.WithContext(ctx).Preload("Role").First(&user, id).Error; err != nil {
		return nil, nil, err
	}
	offset := (pf.Page - 1) * pf.PageSize

	attendancequery := s.db.WithContext(ctx).Table("attendance a").
		Select(`
			a.id AS id,
			u.id AS user_id,
			u.name AS name,
			u.gender AS gender,
			c.id AS company_id,
			c.name AS company_name,
			r.id AS role_id,
			r.display_name AS role_name,
			a.check_date AS check_date,
			a.status AS status
		`).
		Joins("LEFT JOIN user u ON u.id = a.user_id").
		Joins("LEFT JOIN company c ON c.id = u.company_id").
		Joins("LEFT JOIN role r ON r.id = u.role_id")

	attendancequery = attendancequery.Order("a.id DESC")
	attendancequery = applyAccessFilterAttendance(attendancequery, s.db, user.Role, user)
	attendancequery = applyCommonFilterAttendance(attendancequery, filter)

	var totalCount int64
	countQuery := attendancequery.Session(&gorm.Session{})
	if err := countQuery.Count(&totalCount).Error; err != nil {
		return nil, nil, err
	}
	if err := attendancequery.Offset(offset).Limit(pf.PageSize).Scan(&attendance).Error; err != nil {
		return nil, nil, err
	}

	for i := range attendance {
		attendance[i].GenderString = helper.Gender(attendance[i].Gender)
		attendance[i].CheckDate = helper.FormatDate(attendance[i].CheckDate)
	}

	if len(attendance) == 0 {
		return attendance, helper.BuildPaginationMeta(pf, totalCount), nil
	}

	attendanceIDs := make([]int, len(attendance))
	attendanceIndexByID := make(map[int]int, len(attendance))
	for i, a := range attendance {
		attendanceIDs[i] = a.ID
		attendanceIndexByID[a.ID] = i
	}

	var attendancerecords []response.AttendanceRecordResponse

	attendancerecordquery := s.db.WithContext(ctx).Table("attendance_record ar").
		Select(`
			ar.id AS id,
			ar.attendance_id AS attendance_id,
			s.day AS day,
			s.check_in1 AS check_in1,
			s.check_out1 AS check_out1,
			s.check_in2 AS check_in2,
			s.check_out2 AS check_out2,
			at.id AS attendance_type,
			at.name AS attendance_type_name,
			ar.resean AS reason,
			ar.check_time AS check_time,
			ar.type AS type,
			ar.inzone AS inzone,
			ar.latitdude AS latitdude,
			ar.longitude AS longitude
		`).
		Joins("LEFT JOIN shift s ON s.id = ar.shift_id").
		Joins("LEFT JOIN attendance_type at ON at.id = ar.attendance_type").
		Where("ar.attendance_id IN ?", attendanceIDs)

	if err := attendancerecordquery.Scan(&attendancerecords).Error; err != nil {
		return nil, nil, err
	}

	reasonsByAttendanceID := make(map[int][]string, len(attendance))

	for i := range attendancerecords {
		r := &attendancerecords[i]

		scheduledTime := helper.DetermineScheduledTime(r.Type, r.CheckIn1, r.CheckOut1, r.CheckIn2, r.CheckOut2)
		isCheckIn := r.Type == 1 || r.Type == 3
		diff := helper.CalcTimeDiff(r.CheckTime, scheduledTime, isCheckIn)

		idx, ok := attendanceIndexByID[r.AttendanceID]
		if !ok {
			continue
		}

		switch r.Type {
		case 1:
			attendance[idx].CheckIn1 = r.CheckTime
			attendance[idx].CheckIn1Diff = diff
		case 2:
			attendance[idx].CheckOut1 = r.CheckTime
			attendance[idx].CheckOut1Diff = diff
		case 3:
			attendance[idx].CheckIn2 = r.CheckTime
			attendance[idx].CheckIn2Diff = diff
		case 4:
			attendance[idx].CheckOut2 = r.CheckTime
			attendance[idx].CheckOut2Diff = diff
		}

		if r.Reason != "" {
			reasonsByAttendanceID[r.AttendanceID] = append(reasonsByAttendanceID[r.AttendanceID], r.Reason)
		}
	}

	for i := range attendance {
		if reasons, ok := reasonsByAttendanceID[attendance[i].ID]; ok {
			attendance[i].Reason = strings.Join(reasons, "; ")
		}
	}

	return attendance, helper.BuildPaginationMeta(pf, totalCount), nil
}

func (s *attendanceservice) DeleteAttendance(ctx context.Context, id int) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {

		if err := tx.
			Where("attendance_id = ?", id).
			Delete(&model.AttendanceRecord{}).Error; err != nil {

			return fmt.Errorf("failed to delete attendance record: %w", err)
		}

		if err := tx.
			Where("id = ?", id).
			Delete(&model.Attendance{}).Error; err != nil {

			return fmt.Errorf("failed to delete attendance: %w", err)
		}

		return nil
	})
}
