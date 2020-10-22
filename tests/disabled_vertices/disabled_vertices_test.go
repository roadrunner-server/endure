package disabled_vertices

import (
	"testing"

	"github.com/spiral/endure"
)

func TestVertexDisabled(t *testing.T) {
	cont, err := endure.NewContainer(endure.DebugLevel, endure.RetryOnFail(true))
	if err != nil {
		t.Fatal(err)
	}

	_ = cont

}
