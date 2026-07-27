package request

type AttendanceRequestCreate struct {
	CompanyID    int    `json:"company_id"`
	Latitude     string `json:"latitude"`
	Longitude    string `json:"longitude"`
	Reason       string `json:"reason"`
	IsPermission bool   `json:"is_permission"`
}
