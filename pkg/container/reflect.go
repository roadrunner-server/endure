package endure

import (
	"path/filepath"
	"reflect"
	"runtime"
	"strings"

	"github.com/spiral/errors"
)

func providersReturnType(m interface{}) ([]reflect.Type, error) {
	const op = errors.Op("endure_providers_return_type")
	r := reflect.TypeOf(m)
	if r.Kind() != reflect.Func {
		return nil, errors.E(op, errors.ArgType, errors.Errorf("unable to reflect `%s`, expected func. tip: provide function, not structre. for example v.Logger", r.String()))
	}
	ret := make([]reflect.Type, 0, r.NumOut())
	for i := 0; i < r.NumOut(); i++ {
		ret = append(ret, r.Out(i))
	}
	// should be at least 2 parameters
	// error --> nil (hope)
	// type --> initialized
	if r.NumOut() == 0 {
		return nil, errors.E(op, errors.ArgType, errors.Errorf("provider should return at least 1 parameter, but returns `%d`", r.NumOut()))
	}
	// return type, w/o error
	return ret, nil
}

func fnIn(m interface{}) ([]reflect.Type, error) {
	const op = errors.Op("fn_in")
	r := reflect.TypeOf(m)
	if r.Kind() != reflect.Func {
		return nil, errors.E(op, errors.ArgType, errors.Errorf("unable to reflect `%s`, expected func", r.String()))
	}

	out := make([]reflect.Type, 0)
	for i := 0; i < r.NumIn(); i++ {
		out = append(out, r.In(i))
	}

	return out, nil
}

func functionParameters(r reflect.Method) []reflect.Type {
	args := make([]reflect.Type, 0)
	// NumIn returns a function type's input parameter count.
	// It panics if the type's Kind is not Func.
	for i := 0; i < r.Type.NumIn(); i++ {
		// In returns the type of a function type's i'th input parameter.
		// It panics if the type's Kind is not Func.
		// It panics if i is not in the range [0, NumIn()).
		args = append(args, r.Type.In(i))
	}

	return args
}

func getFunctionName(i interface{}) string {
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
