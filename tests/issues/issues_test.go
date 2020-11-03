package issues

import (
	"fmt"
	"testing"
	"time"

	"github.com/spiral/endure"
	"github.com/spiral/endure/tests/issues/issue33"
	issue55_p1 "github.com/spiral/endure/tests/issues/issue55/plugin1"
	issue55_p2 "github.com/spiral/endure/tests/issues/issue55/plugin2"
	issue55_p3 "github.com/spiral/endure/tests/issues/issue55/plugin3"
	"github.com/stretchr/testify/assert"
)

// Provided structure instead of function
func TestEndure_Issue33(t *testing.T) {
	c, err := endure.NewContainer(endure.DebugLevel, nil, endure.RetryOnFail(true))
	assert.NoError(t, err)

	assert.Error(t, c.Register(&issue33.Plugin1{}))
	assert.NoError(t, c.Register(&issue33.Plugin2{}))
}

// https://github.com/spiral/endure/issues/55
// Plugin2 froze execution
// Call Stop on the container
// Should be only 1 stop
func TestEndure_Issue55(t *testing.T) {
	container, err := endure.NewContainer(endure.DebugLevel, nil, endure.RetryOnFail(false))
	assert.NoError(t, err)

	assert.NoError(t, container.Register(&issue55_p1.Plugin1{}))
	assert.NoError(t, container.Register(&issue55_p2.Plugin2{}))
	assert.NoError(t, container.Register(&issue55_p3.Plugin3{}))

	assert.NoError(t, container.Init())

	stopCh := make(chan struct{}, 2)

	go func() {
		time.Sleep(time.Second)
		stopCh <- struct{}{}
	}()

	resCh, err := container.Serve()
	if err == nil {
		t.Fatal(err)
	}

	for {
		select {
		case e := <-resCh:
			fmt.Println(e)
			// first stop
			stopCh <- struct{}{}
			// at the same moment second stop
			err = container.Stop()
			if err != nil {
				t.Fatal(err)
			}
			return
		case <-stopCh:
			err = container.Stop()
			if err != nil {
				t.Fatal(err)
			}
			return
		}
	}
}
