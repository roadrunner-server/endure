package backoff

import (
	"sync"
	"testing"
	"time"

	"github.com/roadrunner-server/endure/v2"
	"github.com/roadrunner-server/endure/v2/tests/backoff/plugins/plugin2"
	"github.com/roadrunner-server/endure/v2/tests/backoff/plugins/plugin3"
	"github.com/roadrunner-server/endure/v2/tests/backoff/plugins/plugin4"
	"github.com/stretchr/testify/assert"
)

func TestEndure_MainThread_Serve_Backoff(t *testing.T) {
	c := endure.New()

	assert.NoError(t, c.Register(&plugin4.Plugin4{}))
	assert.NoError(t, c.Init())

	res, err := c.Serve()
	assert.NoError(t, err)

	wg := &sync.WaitGroup{}

	wg.Add(1)
	go func() {
		for r := range res {
			assert.Error(t, r.Error)
			wg.Done()
		}
	}()
	wg.Wait()
	assert.Error(t, c.Stop())
}

func TestEndure_MainThread_Init_Backoff(t *testing.T) {
	c := endure.New(endure.SetBackoff(time.Second, time.Second*10))

	assert.NoError(t, c.Register(&plugin3.Plugin3{}))
	assert.NoError(t, c.Init())

	res, err := c.Serve()
	assert.NoError(t, err)

	wg := &sync.WaitGroup{}

	now := time.Now().Second()
	wg.Add(1)
	go func() {
		for r := range res {
			if r.Error != nil {
				assert.Error(t, c.Stop())
				wg.Done()
			}
		}
	}()
	wg.Wait()

	after := time.Now().Second()
	// after - now should not be more than 11 as we set in NewContainer
	assert.Greater(t, 11, after-now)
}

func TestEndure_BackoffTimers(t *testing.T) {
	c := endure.New(endure.SetBackoff(time.Second, time.Second*5))

	assert.NoError(t, c.Register(&plugin2.Plugin2{}))
	assert.Error(t, c.Init())

	_, err := c.Serve()
	assert.NoError(t, err)

	assert.NoError(t, c.Stop())
}
