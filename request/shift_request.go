package request

type ShiftUpdateItem struct {
	ID        int     `json:"id"`
	CheckIn1  *string `json:"check_in1"`
	CheckOut1 *string `json:"check_out1"`
	CheckIn2  *string `json:"check_in2"`
	CheckOut2 *string `json:"check_out2"`
	ShiftType *int    `json:"shift_type"`
	IsDayoff  *bool   `json:"is_dayoff"`
}

type ShiftRequestUpdate struct {
	Shifts []ShiftUpdateItem `json:"shifts"`
}

type ShiftCreateItem struct {
	UserID    *int    `json:"user_id"`
	CheckIn1  *string `json:"check_in1"`
	CheckOut1 *string `json:"check_out1"`
	CheckIn2  *string `json:"check_in2"`
	CheckOut2 *string `json:"check_out2"`
	ShiftType *int    `json:"shift_type"`
	Day       *int    `json:"day"`
	IsDayoff  *bool   `json:"is_dayoff"`
}

type ShiftRequestCreate struct {
	Shifts []ShiftCreateItem `json:"shifts"`
}
