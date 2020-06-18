package cascade_test

import (
	"testing"

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

	for {
		select {
		case r := <-res:
			if r.ErrCh != nil {
				err := c.Stop()
				if err != nil {
					t.Fatal(err)
				}
				return
			} else {
				c.Stop()
			}
		}
	}
}

func TestCascade_Init_Err(t *testing.T) {
	c, err := cascade.NewContainer(cascade.TraceLevel)
	assert.NoError(t, err)

	assert.NoError(t, c.Register(&foo4.S4{}))
	assert.NoError(t, c.Register(&foo2.S2{}))
	assert.NoError(t, c.Register(&foo3.S3{}))
	assert.NoError(t, c.Register(&foo1.S1Err{})) // should produce an error during the Init
	assert.NoError(t, c.Init())                    // <-- HERE

	res := c.Serve()

	for {
		select {
		case r := <-res:
			if r.ErrCh != nil {
				err := c.Stop()
				if err != nil {
					assert.Error(t, <-r.ErrCh)
				}
				return
			} else {
				err := c.Stop()
				if err != nil {
					t.Fatal(err)
				}
			}
		}
	}
}
