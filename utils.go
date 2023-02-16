package endure

// Handle all primitive (basic) types
func isPrimitive(str string) bool {
	switch str {
	case "bool":
		return true
	case "string":
		return true
	case "int":
		return true
	case "int8":
		return true
	case "int16":
		return true
	case "int32":
		return true
	case "int64":
		return true
	case "uint":
		return true
	case "uint8":
		return true
	case "uint16":
		return true
	case "uint32":
		return true
	case "uint64":
		return true
	case "uintptr":
		return true
	case "byte":
		return true
	case "rune":
		return true
	case "float32":
		return true
	case "float64":
		return true
	case "complex64":
		return true
	case "complex128":
		return true
	default:
		return false
	}
}
