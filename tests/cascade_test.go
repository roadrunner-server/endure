package cascade_test

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/spiral/cascade"
	"github.com/spiral/cascade/tests/foo1"
	"github.com/spiral/cascade/tests/foo2"
	"github.com/spiral/cascade/tests/foo3"
	"github.com/spiral/cascade/tests/foo4"
)

func TestCascade_Init_OK(t *testing.T) {
	c, err := cascade.NewContainer(cascade.TraceLevel)
	assert.NoError(t, err)

	assert.NoError(t, c.Register(&foo4.S4{}))
	assert.NoError(t, c.Register(&foo2.S2{}))
	assert.NoError(t, c.Register(&foo3.S3{}))
	assert.NoError(t, c.Register(&foo1.S1{}))
	assert.NoError(t, c.Init())

	res := c.Serve()

	go func() {
		for {
			select {
			case r := <-res:
				if r.Err != nil {
					t.Fatal(r.Err)
					return
				}
			}
		}
	}()

	time.Sleep(time.Second * 5)

	assert.NoError(t, c.Stop())
}

func TestCascade_Init_Err(t *testing.T) {
	c, err := cascade.NewContainer(cascade.TraceLevel)
	assert.NoError(t, err)

	assert.NoError(t, c.Register(&foo4.S4{}))
	assert.NoError(t, c.Register(&foo2.S2{}))
	assert.NoError(t, c.Register(&foo3.S3{}))
	assert.NoError(t, c.Register(&foo1.S1Err{})) // should produce an error during the Serve
	assert.NoError(t, c.Init())

	res := c.Serve()

	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		for {
			select {
			case r := <-res:
				println(r.Err.Error())
				assert.Error(t, r.Err)
				assert.NoError(t, c.Stop())
				wg.Done()
				return
			}
		}
	}()

	wg.Wait()
}
