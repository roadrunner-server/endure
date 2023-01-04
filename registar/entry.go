package registar

import (
	"reflect"
)

type returnedType struct {
	retType reflect.Type
	value   func() reflect.Value
	// methods, which used for the providers
	method string
}

type registarEntry struct {
	// plugin type + all provided types
	returnedTypes []*returnedType
	// plugin value
	plugin any
	// weight
	weight uint
}

func (re *registarEntry) Plugin() any {
	return re.plugin
}

func (re *registarEntry) Weight() uint {
	return re.weight
}
