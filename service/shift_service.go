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
	if len(input.Id) == 0 {
		return errors.New("no shift IDs provided")
	}
	if len(input.Id) != len(input.CheckIn1) ||
		len(input.Id) != len(input.CheckOut1) ||
		len(input.Id) != len(input.IsHalft) ||
		len(input.Id) != len(input.IsDayoff) {
		return errors.New("shift fields must have equal length")
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

	for i, id := range input.Id {
		updates := map[string]interface{}{
			"check_in1":  input.CheckIn1[i],
			"check_out1": input.CheckOut1[i],
			"is_halft":   input.IsHalft[i],
			"is_dayoff":  input.IsDayoff[i],
			"check_in2":  input.CheckIn2[i],
			"check_out2": input.CheckOut2[i],
		}
		if err := tx.Model(&model.Shift{}).Where("id = ?", id).Updates(updates).Error; err != nil {
			return err
		}
	}

	if err := tx.Commit().Error; err != nil {
		return err
	}
	committed = true
	return nil
}
