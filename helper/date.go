package helper

import "time"

func CurrentDate() string {
	return time.Now().Format("2006-01-02")
}

func CurrentTime() string {
	return time.Now().Format("15:04:05")
}
