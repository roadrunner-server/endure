package disabled_vertices

import (
	"testing"
	"time"

	"github.com/spiral/endure"
	"github.com/spiral/endure/tests/disabled_vertices/plugin1"
	"github.com/spiral/endure/tests/disabled_vertices/plugin2"
	"github.com/spiral/endure/tests/disabled_vertices/plugin3"
	"github.com/spiral/endure/tests/disabled_vertices/plugin4"
	"github.com/spiral/endure/tests/disabled_vertices/plugin5"
	"github.com/spiral/endure/tests/disabled_vertices/plugin6"
	"github.com/spiral/endure/tests/disabled_vertices/plugin7"
	"github.com/spiral/endure/tests/disabled_vertices/plugin8"
	"github.com/spiral/endure/tests/disabled_vertices/plugin9"
	"github.com/stretchr/testify/assert"
)

func TestVertexDisabled(t *testing.T) {
	cont, err := endure.NewContainer(endure.DebugLevel, endure.RetryOnFail(true))
	if err != nil {
		t.Fatal(err)
	}

	err = cont.Register(&plugin1.Plugin1{}) // disabled
	if err != nil {
		t.Fatal(err)
	}

	err = cont.Register(&plugin2.Plugin2{}) // depends via init
	if err != nil {
		t.Fatal(err)
	}

	err = cont.Init()
	if err != nil {
		t.Fatal()
	}

	_, err = cont.Serve()
	assert.Error(t, err)
}

func TestDisabledViaInterface(t *testing.T) {
	cont, err := endure.NewContainer(endure.DebugLevel, endure.RetryOnFail(true))
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
	if err != nil {
		t.Fatal()
	}

	errCh, err := cont.Serve()
	assert.NoError(t, err)

	tt := time.NewTicker(time.Second)
	// should be one vertex
	for {
		select {
		case e := <-errCh:
			assert.NoError(t, e.Error)
		case <-tt.C:
			assert.NoError(t, cont.Stop())
			tt.Stop()
			return
		}
	}
}

func TestDisabledRoot(t *testing.T) {
	cont, err := endure.NewContainer(endure.DebugLevel, endure.RetryOnFail(true))
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
	if err != nil {
		t.Fatal()
	}

	_, err = cont.Serve() // no plugins to run
	assert.Error(t, err)
}
