package helper

func ShiftType(shifttype int) string {
	switch shifttype {
	case 1:
		return "ធ្វេីការពេញម៉ោង"
	case 2:
		return "ធ្វេីការតែមួយព្រឹក"
	case 3:
		return "ធ្វេីការតែមួយរសៀល"
	default:
		return ""
	}
}
