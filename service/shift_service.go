package service

import (
	"context"
	"errors"
	"mysql/config"
	"mysql/model"
	"mysql/request"

	"gorm.io/gorm"
)

type ShiftService interface {
	UpdateShift(ctx context.Context, input request.ShiftRequestUpdate) error
}

type shiftservice struct {
	db *gorm.DB
}

func NewShiftService() ShiftService {
	return &shiftservice{
		db: config.DB,
	}
}

func (s *shiftservice) UpdateShift(ctx context.Context, input request.ShiftRequestUpdate) error {
	if len(input.Shifts) == 0 {
		return errors.New("no shifts provided")
	}

	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, shift := range input.Shifts {
			if err := tx.Model(&model.Shift{}).
				Where("id = ?", shift.ID).
				Updates(map[string]interface{}{
					"check_in1":  &shift.CheckIn1,
					"check_out1": &shift.CheckOut1,
					"check_in2":  &shift.CheckIn2,
					"check_out2": &shift.CheckOut2,
					"is_halft":   &shift.IsHalft,
					"is_dayoff":  &shift.IsDayoff,
				}).Error; err != nil {
				return err
			}
		}
		return nil
	})
}
