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
