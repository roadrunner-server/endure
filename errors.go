package cascade

import (
	"fmt"
	"runtime/debug"

	"github.com/pkg/errors"
)

type Error struct {
	Err   error
	Code  int
	Stack []byte
}

var FailedToSortTheGraph = Error{
	Err:   errors.New("failed to topologically sort the graph"),
	Code:  500,
	Stack: debug.Stack(),
}

var ErrorDuringInit = Error{
	Err:   errors.New("error during the Init function call"),
	Code:  501,
	Stack: debug.Stack(),
}

var FailedToGetTheVertex = Error{
	Err:   errors.New("failed to get vertex from the graph, vertex is nil"),
	Code:  502,
	Stack: debug.Stack(),
}

var BackoffRetryError = Error{
	Err:   errors.New("backoff finished with error"),
	Code:  503,
	Stack: debug.Stack(),
}

var ErrorDuringServe = Error{
	Err:   errors.New("error during the Serve function call"),
	Code:  504,
	Stack: debug.Stack(),
}

var typeNotImplementError = errors.New("type should implement Service interface")
var vertexAlreadyExists = func(name string) error { return fmt.Errorf("vertex `%s` already exists", name) }
var unknownErrorOccurred = errors.New("unknown error occurred during the function call")
