package dep

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
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

	assert.Equal(t, reflect.Interface, tt.Type.Kind())
	assert.Equal(t, "F", tt.Method)
}
