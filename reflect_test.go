package cascade

import (
	"log"
	"testing"
)

func TestReturnType(t *testing.T) {
	log.Print(returnType(func() string {
		return "hello"
	}))
}

func TestArgType(t *testing.T) {
	log.Print(argType(func(string, int) string {
		return "hello"
	}))
}
