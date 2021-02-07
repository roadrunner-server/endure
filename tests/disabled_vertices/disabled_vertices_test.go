package disabled_vertices

import (
	"testing"

	endure "github.com/spiral/endure/pkg/container"
	"github.com/spiral/endure/tests/disabled_vertices/plugin3"
	"github.com/spiral/endure/tests/disabled_vertices/plugin4"
	"github.com/spiral/endure/tests/disabled_vertices/plugin5"
	"github.com/spiral/endure/tests/disabled_vertices/plugin6"
	"github.com/spiral/endure/tests/disabled_vertices/plugin7"
	"github.com/spiral/endure/tests/disabled_vertices/plugin8"
	"github.com/spiral/endure/tests/disabled_vertices/plugin9"
	"github.com/stretchr/testify/assert"
)

// TODO tests temporarily disabled until proper disable will be implemented
// func TestVertexDisabled(t *testing.T) {
//	cont, err := endure.NewContainer(nil, endure.RetryOnFail(true))
//	if err != nil {
//		t.Fatal(err)
//	}
//
//	err = cont.Register(&plugin1.Plugin1{}) // disabled
//	if err != nil {
//		t.Fatal(err)
//	}
//
//	err = cont.Register(&plugin2.Plugin2{}) // depends via init
//	if err != nil {
//		t.Fatal(err)
//	}
//
//	err = cont.Init()
//	if err != nil {
//		t.Fatal()
//	}
//
//	_, err = cont.Serve()
//	assert.Error(t, err)
// }

func TestDisabledViaInterface(t *testing.T) {
	cont, err := endure.NewContainer(nil, endure.RetryOnFail(true))
	if err != nil {
		t.Fatal(err)
	}

	err = cont.Register(&plugin3.Plugin3{}) // disabled
	if err != nil {
		t.Fatal(err)
	}

	err = cont.Register(&plugin4.Plugin4{}) // depends via init
	if err != nil {
		t.Fatal(err)
	}

	err = cont.Register(&plugin5.Plugin5{}) // depends via init
	if err != nil {
		t.Fatal(err)
	}

	err = cont.Init()
	assert.Error(t, err)
}

// TODO tests temporarily disabled until proper disable will be implemented
func TestDisabledRoot(t *testing.T) {
	cont, err := endure.NewContainer(nil)
	if err != nil {
		t.Fatal(err)
	}

	err = cont.Register(&plugin6.Plugin6{}) // Root
	if err != nil {
		t.Fatal(err)
	}

	err = cont.Register(&plugin7.Plugin7{}) // should be disabled
	if err != nil {
		t.Fatal(err)
	}

	err = cont.Register(&plugin8.Plugin8{}) // should be disabled
	if err != nil {
		t.Fatal(err)
	}

	err = cont.Register(&plugin9.Plugin9{}) // should be disabled
	if err != nil {
		t.Fatal(err)
	}

	err = cont.Init()
	assert.Error(t, err)
}
