package helper

import "time"

func GetCurrentDay() int {
	day := int(time.Now().Weekday())
	if day == 0 {
		return 7
	}
	return day
}
