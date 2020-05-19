package cascade

import (
	"log"
	"testing"
)

func TestReturnType(t *testing.T) {
	log.Print(returnKind(func() string {
		return "hello"
	}))
}
