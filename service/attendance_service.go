package service

import (
	"context"
	"errors"
	"fmt"
	"mysql/config"
	"mysql/helper"
	"mysql/model"
	"mysql/request"
	"mysql/response"
	"mysql/utils"
	"strconv"
	"time"

	"gorm.io/gorm"
)

type AttendanceService interface {
	CreateAttendance(ctx context.Context, id int, input request.AttendanceRequestCreate) error
	GetAttendance(ctx context.Context, id int, pf request.Pagination, filter map[string]string) ([]response.AttendanceResponse, *model.PaginationMetadata, error)
}

type attendanceservice struct {
	db *gorm.DB
}

func NewAttendanceService() AttendanceService {
	return &attendanceservice{
		db: config.DB,
	}
}

func (s *attendanceservice) CreateAttendance(ctx context.Context, id int, input request.AttendanceRequestCreate) error {
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

	now := time.Now()
	currentDate := now.Format("2006-01-02")
	currentTime := now.Format("15:04:05")
	dayofweek := helper.GetCurrentDay()

	var user model.User
	if err := s.db.WithContext(ctx).Preload("Company").First(&user, id).Error; err != nil {
		tx.Rollback()
		return err
	}

	var shift model.Shift
	if err := tx.Where("user_id = ? AND day = ?", user.ID, dayofweek).First(&shift).Error; err != nil {
		tx.Rollback()
		return err
	}
	if shift.IsDayoff {
		tx.Rollback()
		return errors.New("today is dayoff")
	}

	companyLat, err := strconv.ParseFloat(user.Company.Latitude, 64)
	if err != nil {
		tx.Rollback()
		return err
	}
	companyLng, err := strconv.ParseFloat(user.Company.Longitude, 64)
	if err != nil {
		tx.Rollback()
		return err
	}
	userLat, err := strconv.ParseFloat(input.Latitude, 64)
	if err != nil {
		tx.Rollback()
		return err
	}
	userLng, err := strconv.ParseFloat(input.Longitude, 64)
	if err != nil {
		tx.Rollback()
		return err
	}
	radius, err := strconv.ParseFloat(user.Company.Radius, 64)
	if err != nil {
		tx.Rollback()
		return err
	}

	distance := utils.CalculateDistance(companyLat, companyLng, userLat, userLng)
	inZone := distance <= radius

	var attendance model.Attendance
	err = tx.Where("user_id = ? AND check_date = ?", user.ID, currentDate).First(&attendance).Error
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			tx.Rollback()
			return err
		}

		attendance = model.Attendance{
			UserID:    user.ID,
			CheckDate: currentDate,
			Status:    "WORKING",
		}
		if err := tx.Create(&attendance).Error; err != nil {
			tx.Rollback()
			return err
		}
	}

	var existingRecords []model.AttendanceRecord
	if err := tx.Where("attendance_id = ?", attendance.ID).
		Order("id ASC").
		Find(&existingRecords).Error; err != nil {
		tx.Rollback()
		return err
	}

	recordCount := len(existingRecords)

	// Build session list based on ShiftType
	// ShiftType 1 = Full day    → CheckIn1, CheckOut1, CheckIn2, CheckOut2
	// ShiftType 2 = Morning only → CheckIn1, CheckOut1
	// ShiftType 3 = Evening only → CheckIn2, CheckOut2

	type sessionConfig struct {
		scheduledTime string
		isCheckIn     bool
		recordType    int // 1=CheckIn1, 2=CheckOut1, 3=CheckIn2, 4=CheckOut2
	}

	var sessions []sessionConfig

	switch shift.ShiftType {

	case 2:
		if shift.CheckIn1 == nil || shift.CheckOut1 == nil {
			return errors.New("ធ្វេីការវែនព្រឹកតែមិនទាន់ដាក់ម៉ោងចេញចូល")
		}
		sessions = []sessionConfig{
			{scheduledTime: *shift.CheckIn1, isCheckIn: true, recordType: 1},
			{scheduledTime: *shift.CheckOut1, isCheckIn: false, recordType: 2},
		}
	case 3:
		if shift.CheckIn2 == nil || shift.CheckOut2 == nil {
			return errors.New("ធ្វេីការវែនល្ងាចតែមិនទាន់ដាក់ម៉ោងចេញចូល")
		}
		sessions = []sessionConfig{
			{scheduledTime: *shift.CheckIn2, isCheckIn: true, recordType: 3},
			{scheduledTime: *shift.CheckOut2, isCheckIn: false, recordType: 4},
		}
	default:
		if shift.CheckIn1 == nil || shift.CheckOut1 == nil || shift.CheckIn2 == nil || shift.CheckOut2 == nil {
			return errors.New("គ្មានម៉ោងធ្វេីការ")
		}
		sessions = []sessionConfig{
			{scheduledTime: *shift.CheckIn1, isCheckIn: true, recordType: 1},
			{scheduledTime: *shift.CheckOut1, isCheckIn: false, recordType: 2},
			{scheduledTime: *shift.CheckIn2, isCheckIn: true, recordType: 3},
			{scheduledTime: *shift.CheckOut2, isCheckIn: false, recordType: 4},
		}
	}

	maxRecords := len(sessions)

	if recordCount >= maxRecords {
		if err := tx.Model(&model.Attendance{}).
			Where("id = ?", attendance.ID).
			Update("status", "COMPLETE").Error; err != nil {
			return fmt.Errorf("failed to update attendance status: %w", err)
		}
		if err := tx.Commit().Error; err != nil {
			return err
		}
		committed = true
		return errors.New("all check-ins and check-outs completed for today")
	}

	current := sessions[recordCount]

	attendanceType := helper.DetermineAttendanceType(currentTime, current.scheduledTime, current.isCheckIn)

	record := model.AttendanceRecord{
		AttendanceID:   attendance.ID,
		ShiftID:        shift.ID,
		AttendanceType: attendanceType,
		Reason:         input.Reason,
		CheckTime:      currentTime,
		Type:           current.recordType,
		Inzone:         inZone,
	}
	if err := tx.Create(&record).Error; err != nil {
		return err
	}

	if recordCount+1 >= maxRecords {
		if err := tx.Model(&model.Attendance{}).
			Where("id = ?", attendance.ID).
			Update("status", "COMPLETE").Error; err != nil {
			return fmt.Errorf("failed to update attendance status: %w", err)
		}
	}

	if err := tx.Commit().Error; err != nil {
		return err
	}
	committed = true
	return nil

}

func applyAccessFilterAttendance(query *gorm.DB, db *gorm.DB, role model.Role, user model.User) *gorm.DB {
	if role.Level > 1 && role.Level < 7 {
		return query.Where("u.company_id = ?", user.CompanyID)
	} else if role.Level <= 1 {
		return query.Where("u.id =?", user.ID)
	}

	return query
}

func applyCommonFilterAttendance(query *gorm.DB, filter map[string]string) *gorm.DB {
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
		case "check_date":
			query = query.Where("a.check_date =?", value)
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
        ar.inzone AS inzone
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
