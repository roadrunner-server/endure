package disabled_vertices

import (
	"testing"

	"github.com/spiral/endure"
	"github.com/spiral/endure/tests/disabled_vertices/plugin1"
	"github.com/spiral/endure/tests/disabled_vertices/plugin2"
)

func TestVertexDisabled(t *testing.T) {
	cont, err := endure.NewContainer(endure.DebugLevel, endure.RetryOnFail(true))
	if err != nil {
		t.Fatal(err)
	}

	err = cont.Register(&plugin1.Plugin1{})
	if err != nil {
		t.Fatal(err)
	}

	err = cont.Register(&plugin2.Plugin2{})
	if err != nil {
		t.Fatal(err)
	}

	err = cont.Init()
	if err != nil {
		t.Fatal()
	}
}
