package tool

func IsValidInt(v interface{}) bool {
	switch n := v.(type) {
	case int:
		return n > 0
	case int32:
		return n > 0
	case int64:
		return n > 0
	default:
		return false
	}
}
