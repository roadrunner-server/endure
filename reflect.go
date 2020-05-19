package cascade

import (
	"fmt"
	"reflect"
)

func returnKind(m interface{}) (reflect.Type, error) {
	r := reflect.TypeOf(m)
	if r.Kind() != reflect.Func {
		return nil, fmt.Errorf("unable to reflect `%s`, expected func", r.String())
	}

	return nil, nil
}

func argumentKind(m interface{}) ([]reflect.Type, error) {
	return nil, nil
}
