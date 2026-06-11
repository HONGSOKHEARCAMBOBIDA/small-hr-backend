package helper

import "strconv"

func ParseFloat(s string) float64 {
	if s == "" {
		return 0
	}
	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0
	}
	return v
}

func FormatFloat(f float64) string {
	return strconv.FormatFloat(f, 'f', 2, 64)
}
