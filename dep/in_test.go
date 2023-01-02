package dep

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

type BarBaz interface {
	Bar()
	b()
}

func TestImplements(t *testing.T) {
	in := Fits(func(p any) {
		println("foo")
	}, (*BarBaz)(nil))

	assert.Equal(t, reflect.Interface, in.Type.Kind())
}
