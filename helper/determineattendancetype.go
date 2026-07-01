package helper

import "time"

func DetermineAttendanceType(actual, scheduled string, isCheckIn bool) int {
	const gracePeriod = 30 * time.Second

	actualT, err1 := time.Parse("15:04:05", actual)
	scheduledT, err2 := time.Parse("15:04:05", scheduled)
	if err1 != nil || err2 != nil {

		if isCheckIn {
			return 2
		}
		return 5
	}

	diff := actualT.Sub(scheduledT)

	if isCheckIn {
		switch {
		case diff < -gracePeriod:
			return 1
		case diff <= gracePeriod:
			return 2
		default:
			return 3
		}
	}

	switch {
	case diff < -gracePeriod:
		return 4
	case diff <= gracePeriod:
		return 5
	default:
		return 6
	}
}
