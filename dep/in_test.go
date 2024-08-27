package dep

import (
	"reflect"
	"testing"
)

type BarBaz interface {
	Bar()
	b()
}

func TestImplements(t *testing.T) {
	in := Fits(func(p any) {
		println("foo")
	}, (*BarBaz)(nil))

	if reflect.Interface != in.Type.Kind() {
		t.Fail()
	}
}
