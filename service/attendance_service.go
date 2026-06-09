package service

import (
	"context"
	"errors"
	"fmt"
	"mysql/config"
	"mysql/helper"
	"mysql/model"
	"mysql/request"
	"mysql/utils"
	"strconv"
	"time"

	"gorm.io/gorm"
)

type AttendanceService interface {
	CreateAttendance(ctx context.Context, id int, input request.AttendanceRequestCreate) error
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
	err = tx.Where("user_id = ? AND check_date = ?", user.ID, currentDate).
		First(&attendance).Error
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

	// Determine action: 0=CheckIn1, 1=CheckOut1, 2=CheckIn2, 3=CheckOut2
	// recordCount 0 → CheckIn1
	// recordCount 1 → CheckOut1
	// recordCount 2 → CheckIn2
	// recordCount 3 → CheckOut2

	if recordCount >= 2 && shift.CheckIn2 == "" {
		tx.Rollback()
		return errors.New("no second session in shift")
	}

	if recordCount >= 4 {
		if err := tx.Model(&model.Attendance{}).
			Where("id = ?", attendance.ID).
			Update("status", "COMPLETE").Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to update attendance status: %w", err)
		}
		if err := tx.Commit().Error; err != nil {
			return err
		}
		committed = true
		return errors.New("all check-ins and check-outs completed for today")
	}

	// Determine attendance_type by comparing current time vs shift schedule
	// attendance_type:
	// 1 = ចូលធ្វើការមុនម៉ោង  (check-in early)
	// 2 = ចូលធ្វើការទាន់ម៉ោង (check-in on time)
	// 3 = ចូលធ្វើការយឺត      (check-in late)
	// 4 = ចេញពីធ្វើការមុនម៉ោង (check-out early)
	// 5 = ចេញពីធ្វើការត្រឹមម៉ោង(check-out on time)
	// 6 = ចេញពីធ្វើការក្រោយម៉ោង(check-out overtime)

	var scheduledTime string
	var isCheckIn bool

	switch recordCount {
	case 0:
		scheduledTime = shift.CheckIn1
		isCheckIn = true
	case 1:
		scheduledTime = shift.CheckOut1
		isCheckIn = false
	case 2:
		scheduledTime = shift.CheckIn2
		isCheckIn = true
	case 3:
		scheduledTime = shift.CheckOut2
		isCheckIn = false
	}

	attendanceType := helper.DetermineAttendanceType(currentTime, scheduledTime, isCheckIn)

	record := model.AttendanceRecord{
		AttendanceID:   attendance.ID,
		ShiftID:        shift.ID,
		AttendanceType: attendanceType,
		Reason:         input.Reason,
		CheckTime:      currentTime,
		Type:           recordCount + 1, // 1=CheckIn1, 2=CheckOut1, 3=CheckIn2, 4=CheckOut2
		Inzone:         inZone,
	}
	if err := tx.Create(&record).Error; err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Commit().Error; err != nil {
		return err
	}
	committed = true
	return nil
}
