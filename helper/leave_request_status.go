package helper

func LeaveRequestStatus(status int) string {
	switch status {
	case 1:
		return "PENDING"
	case 2:
		return "APPROVED"
	case 3:
		return "REJECTED"
	case 4:
		return "CANCELLED"
	default:
		return ""
	}
}
