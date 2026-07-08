package request

type AttendanceRequestCreate struct {
	Latitude     string `json:"latitude"`
	Longitude    string `json:"longitude"`
	Reason       string `json:"reason"`
	IsPermission bool   `json:"is_permission"`
}
