package request

type ShiftUpdateItem struct {
	ID        int     `json:"id"`
	CheckIn1  *string `json:"check_in1"`
	CheckOut1 *string `json:"check_out1"`
	CheckIn2  *string `json:"check_in2"`
	CheckOut2 *string `json:"check_out2"`
	IsHalft   *bool   `json:"is_halft"`
	IsDayoff  *bool   `json:"is_dayoff"`
}

type ShiftRequestUpdate struct {
	Shifts []ShiftUpdateItem `json:"shifts"`
}
