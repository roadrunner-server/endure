package cascade

import (
	"fmt"
	"reflect"
)

func returnType(m interface{}) (reflect.Type, error) {
	r := reflect.TypeOf(m)
	if r.Kind() != reflect.Func {
		return nil, fmt.Errorf("unable to reflect `%s`, expected func", r.String())
	}

	if r.NumOut() != 1 {
		return nil, fmt.Errorf("unable to determinate return type of `%s`", r.String())
	}

	return r.Out(0), nil
}

func argType(m interface{}) ([]reflect.Type, error) {
	r := reflect.TypeOf(m)
	if r.Kind() != reflect.Func {
		return nil, fmt.Errorf("unable to reflect `%s`, expected func", r.String())
	}

	out := make([]reflect.Type, 0)
	for i := 0; i < r.NumIn(); i++ {
		out = append(out, r.In(i))
	}

	return out, nil
}

func functionParameters(r reflect.Method) ([]reflect.Type, error) {
	args := make([]reflect.Type, 0)
	// NumIn returns a function type's input parameter count.
	// It panics if the type's Kind is not Func.
	for i := 0; i < r.Type.NumIn(); i++ {
		// In returns the type of a function type's i'th input parameter.
		// It panics if the type's Kind is not Func.
		// It panics if i is not in the range [0, NumIn()).
		args = append(args, r.Type.In(i))
		//println(r.Type.In(i).String())
	}

	return args, nil
}

func typeMatches(r reflect.Type, v interface{}) bool {
	to := reflect.TypeOf(v)

	println(to.String())
	println(r.String())


	if r.ConvertibleTo(to) {
		return true
	}

	return false
}
