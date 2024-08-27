package dep

import (
	"reflect"
	"testing"
)

type FooBar interface {
	Foo()
	Bar()
}

type TestStruct struct{}

func (ts *TestStruct) Foo() {}
func (ts *TestStruct) Bar() {}

type Plugin struct{}

func (p *Plugin) F() (*TestStruct, error) {
	return &TestStruct{}, nil
}

func TestOutType(t *testing.T) {
	p := Plugin{}
	tt := Bind((*FooBar)(nil), p.F)

	if reflect.Interface != tt.Type.Kind() {
		t.Fail()
	}

	if tt.Method != "F" {
		t.Fail()
	}
}
