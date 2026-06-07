package helper

func Gender(gender int) string {
	switch gender {
	case 1:
		return "ប្រុស"
	case 2:
		return "ស្រី"
	default:
		return ""
	}
}
