package request

type AttendanceRequestCreate struct {
	Latitude  string `json:"latitude"`
	Longitude string `json:"longitude"`
	Reason    string `json:"reason"`
}
