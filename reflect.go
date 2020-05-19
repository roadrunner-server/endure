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

func argumentKind(m interface{}) ([]reflect.Type, error) {
	return nil, nil
}
