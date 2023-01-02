package dep

import (
	"reflect"
)

type In struct {
	Type     reflect.Type
	Callback func(any)
}

func Fits(callback func(p any), tp any) *In {
	if reflect.TypeOf(tp) == nil {
		panic("nil type provided, should be of the form of: (*FooBar)(nil), not (FooBar)(nil)")
	}

	if reflect.TypeOf(tp).Elem().Kind() != reflect.Interface {
		panic("type should be of the Interface type")
	}

	return &In{
		Type:     reflect.TypeOf(tp).Elem(),
		Callback: callback,
	}
}
