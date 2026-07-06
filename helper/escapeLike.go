package helper

import "strings"

func EscapeLike(value string) string {
	replacer := strings.NewReplacer("%", "\\%", "_", "\\_")
	return replacer.Replace(value)
}
