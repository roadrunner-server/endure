package cascade

import "errors"

var Disabled = errors.New("service disabled")

type Container interface {
	Get(name string) interface{}
	Fork(section string) Container
}
