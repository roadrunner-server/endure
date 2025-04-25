package dep

import (
	"fmt"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
)

type Out struct {
	Type   reflect.Type
	Method string
}

func Bind(tp any, method any) *Out {
	if reflect.TypeOf(tp) == nil {
		panic("nil type provided, should be of the form of: (*FooBar)(nil), not (FooBar)(nil)")
	}

	if reflect.TypeOf(tp).Elem().Kind() != reflect.Interface {
		panic(fmt.Errorf("provided type should be of the type: Interface, actual type is: %s", reflect.TypeOf(tp).Elem().Kind().String()))
	}

	r := reflect.TypeOf(method)
	if r.Kind() != reflect.Func {
		panic("second argument should be a function")
	}

	for i := range r.NumOut() {
		// skip errors
		if r.Out(i) == reflect.TypeOf((*error)(nil)).Elem() {
			continue
		}

		if !r.Out(i).Implements(reflect.TypeOf(tp).Elem()) {
			panic("provided method should return an implementation of the provided interface")
		}
	}

	if r.NumIn() > 0 {
		panic("dep.Bind function should not receive any arguments")
	}

	return &Out{
		Type:   reflect.TypeOf(tp).Elem(),
		Method: getFunctionName(method),
	}
}

func getFunctionName(i any) string {
	rawName := runtime.FuncForPC(reflect.ValueOf(i).Pointer()).Name()
	name := strings.TrimPrefix(filepath.Ext(rawName), ".")

	/* This seems to have become -fm on tip, by the way.

	Your code is getting a method value. p.beHappy is the beHappy method bound to the specific value of p.
	That is implemented by creating a function closure, and the code for that closure needs a fn.
	The compiler happens to make that fn by sticking fm on the end,
	but it could be anything that won't conflict with any other function fn.
	There isn't any way to fn that function in Go, so the fn is irrelevant for anything other than the debugger or, as you see, FuncForPC.

	In the reflection, we would have this suffix
	*/
	return strings.TrimSuffix(name, "-fm")
}
