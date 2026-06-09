package helper

func TypeFormat(typeint int) string {
	switch typeint {
	case 1:
		return "ចូលធ្វេីការវែនទី១"
	case 2:
		return "ចេញពីធ្វេីការវែនទី១"
	case 3:
		return "ចូលធ្វេីការវែនទី២"
	case 4:
		return "ចេញពីធ្វេីការវែនទី២"
	default:
		return ""
	}
}
