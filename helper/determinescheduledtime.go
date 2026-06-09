package helper

import (
	"fmt"
	"strings"
	"time"
)

// type 1 = CheckIn1, 2 = CheckOut1, 3 = CheckIn2, 4 = CheckOut2
func DetermineScheduledTime(recordType int, checkIn1, checkOut1, checkIn2, checkOut2 string) string {
	switch recordType {
	case 1:
		return checkIn1
	case 2:
		return checkOut1
	case 3:
		return checkIn2
	case 4:
		return checkOut2
	}
	return ""
}

// CalcTimeDiff returns the difference between actual check time and scheduled time.
// Positive = late/overtime, Negative = early.
// Format: "+HH:MM:SS" or "-HH:MM:SS"
func CalcTimeDiff(checkTime, scheduledTime string, isCheckIn bool) string {
	if scheduledTime == "" || checkTime == "" {
		return ""
	}
	actual, err1 := time.Parse("15:04:05", checkTime)
	scheduled, err2 := time.Parse("15:04:05", scheduledTime)
	if err1 != nil || err2 != nil {
		return ""
	}

	diff := actual.Sub(scheduled)
	if diff < 0 {
		diff = -diff
	}

	h := int(diff.Hours())
	m := int(diff.Minutes()) % 60
	s := int(diff.Seconds()) % 60

	parts := ""
	if h > 0 {
		parts += fmt.Sprintf("%d ម៉ោង ", h)
	}
	if m > 0 {
		parts += fmt.Sprintf("%d នាទី ", m)
	}
	if s > 0 && h == 0 {
		parts += fmt.Sprintf("%d វិនាទី", s)
	}
	parts = strings.TrimSpace(parts)

	if parts == "" {
		if isCheckIn {
			return "ចូលធ្វើការទាន់ម៉ោង"
		}
		return "ចេញពីធ្វើការត្រឹមម៉ោង"
	}

	actual2, _ := time.Parse("15:04:05", checkTime)
	scheduled2, _ := time.Parse("15:04:05", scheduledTime)
	diff2 := actual2.Sub(scheduled2)

	if isCheckIn {
		if diff2 < 0 {
			return fmt.Sprintf("មុនម៉ោង %s", parts)
		}
		return fmt.Sprintf("យឺត %s", parts)
	}

	// check-out
	if diff2 < 0 {
		return fmt.Sprintf("មុនម៉ោង %s", parts)
	}
	return fmt.Sprintf("បន្ថែមម៉ោង %s", parts)
}
