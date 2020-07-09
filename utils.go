package cascade

import (
	"reflect"
	"strings"
)

func removePointerAsterisk(s string) string {
	return strings.Trim(s, "*")
}

func isReference(t reflect.Type) bool {
	return t.Kind() == reflect.Ptr
}

// TODO add all primitive types
func isPrimitive(str string) bool {
	switch str {
	case "int":
		return true
	default:
		return false
	}
}
