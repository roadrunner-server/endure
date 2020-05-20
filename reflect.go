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

func argrType(r reflect.Method) ([]reflect.Type, error) {
	out := make([]reflect.Type, 0)
	for i := 0; i < r.Type.NumIn(); i++ {
		out = append(out, r.Type.In(i))
	}

	return out, nil
}

func typeMatches(r reflect.Type, v interface{}) bool {
	to := reflect.TypeOf(v)

	//if to.Implements(r) {
	//	return true
	//}

	if r.ConvertibleTo(to) {
		return true
	}

	return false
}
