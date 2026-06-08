package request

type ShiftRequestUpdate struct {
	Id        []int    `json:"id"`
	CheckIn1  []string `json:"check_in1" gorm:"column:check_in1"`
	CheckOut1 []string `json:"check_out1" gorm:"column:check_out1"`
	CheckIn2  []string `json:"check_in2" gorm:"column:check_in2"`
	CheckOut2 []string `json:"check_out2" gorm:"column:check_out2"`
	IsHalft   []bool   `json:"is_halft" gorm:"column:is_halft"`
	IsDayoff  []bool   `json:"is_dayoff" gorm:"column:is_dayoff"`
}
