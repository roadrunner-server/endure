package cascade

import (
	"fmt"

	"github.com/pkg/errors"
)

var typeNotImplementError = errors.New("type should implement Service interface")
var vertexAlreadyExists = func(name string) error { return fmt.Errorf("vertex `%s` already exists", name) }
var unknownErrorOccurred = errors.New("unknown error occurred during the function call")
