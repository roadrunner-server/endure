package stress

import (
	"sync"
	"testing"

	"github.com/spiral/endure"
	"github.com/spiral/endure/tests/stress/CollectorFuncReturn"
	"github.com/spiral/endure/tests/stress/InitErr"
	"github.com/spiral/endure/tests/stress/ServeErr"
	"github.com/spiral/endure/tests/stress/ServeRetryErr"
	"github.com/spiral/endure/tests/stress/mixed"
	"github.com/spiral/errors"
	"github.com/stretchr/testify/assert"
)

func TestEndure_Init_Err(t *testing.T) {
	c, err := endure.NewContainer(endure.DebugLevel, nil, endure.RetryOnFail(false))
	assert.NoError(t, err)

	assert.NoError(t, c.Register(&InitErr.S1Err{}))
	assert.NoError(t, c.Register(&InitErr.S2Err{})) // should produce an error during the Init
	assert.Error(t, c.Init())
}

func TestEndure_DoubleStop_Err(t *testing.T) {
	c, err := endure.NewContainer(endure.DebugLevel, nil, endure.RetryOnFail(false))
	assert.NoError(t, err)

	assert.NoError(t, c.Register(&InitErr.S1Err{}))
	assert.NoError(t, c.Register(&InitErr.S2Err{})) // should produce an error during the Init
	assert.Error(t, c.Init())
	assert.NoError(t, c.Stop())
	// recognizer: can't transition from state: Stopped by event Stop
	assert.Error(t, c.Stop())
}

func TestEndure_Serve_Err(t *testing.T) {
	c, err := endure.NewContainer(endure.DebugLevel, nil, endure.RetryOnFail(false))
	assert.NoError(t, err)

	assert.NoError(t, c.Register(&ServeErr.S4ServeError{})) // should produce an error during the Serve
	assert.NoError(t, c.Register(&ServeErr.S2{}))
	assert.NoError(t, c.Register(&ServeErr.S3ServeError{}))
	assert.NoError(t, c.Register(&ServeErr.S5{}))
	assert.NoError(t, c.Register(&ServeErr.S1ServeErr{}))
	err = c.Init()
	if err != nil {
		t.Fatal(err)
	}

	_, err = c.Serve()
	assert.Error(t, err)
}

/* The scenario for this test is the following:
time X is 0s
1. After X+1s S2ServeErr produces error in Serve
2. At the same time at X+1s S1Err also produces error in Serve
3. In case of S2ServeErr vertices S5 and S4V should be restarted
4. In case of S1Err vertices S5 -> S4V -> S2ServeErr (with error in Serve in X+5s) -> S1Err should be restarted
*/
func TestEndure_Serve_Retry_Err(t *testing.T) {
	c, err := endure.NewContainer(endure.DebugLevel, nil, endure.RetryOnFail(true))
	assert.NoError(t, err)

	assert.NoError(t, c.Register(&ServeRetryErr.S4{}))
	assert.NoError(t, c.Register(&ServeRetryErr.S2{}))
	assert.NoError(t, c.Register(&ServeRetryErr.S2ServeErr{})) // Random error here
	assert.NoError(t, c.Register(&ServeRetryErr.S3{}))
	assert.NoError(t, c.Register(&ServeRetryErr.S5{}))
	assert.NoError(t, c.Register(&ServeRetryErr.S1ServeErr{})) // should produce an error during the Serve
	assert.NoError(t, c.Init())

	res, err := c.Serve()
	assert.NoError(t, err)

	// we can't be sure, what node will be processed first
	ord := [2]string{"ServeRetryErr.S1ServeErr", "ServeRetryErr.S2ServeErr"}

	count := 0

	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		for r := range res {
			assert.Error(t, r.Error)
			if r.VertexID == ord[0] || r.VertexID == ord[1] {
				count++
				if count == 2 {
					assert.NoError(t, c.Stop())
					wg.Done()
					return
				}
			} else {
				assert.Fail(t, "vertex should be in the ord slice")
				wg.Done()
				return
			}
		}
	}()

	wg.Wait()
}

/* The scenario for this test is the following:
time X is 0s
1. After X+1s S2ServeErr produces error in Serve
2. At the same time at X+1s S1Err also produces error in Serve
3. In case of S2ServeErr vertices S5 and S4V should be restarted
4. In case of S1Err vertices S5 -> S4V -> S2ServeErr (with error in Serve in X+5s) -> S1Err should be restarted
5. Test should receive at least 100 errors
*/
func TestEndure_Serve_Retry_100_Err(t *testing.T) {
	c, err := endure.NewContainer(endure.DebugLevel, nil, endure.RetryOnFail(true))
	assert.NoError(t, err)

	assert.NoError(t, c.Register(&ServeRetryErr.S4{}))
	assert.NoError(t, c.Register(&ServeRetryErr.S2{}))
	assert.NoError(t, c.Register(&ServeRetryErr.S2ServeErr{})) // Random error here
	assert.NoError(t, c.Register(&ServeRetryErr.S3{}))
	assert.NoError(t, c.Register(&ServeRetryErr.S5{}))
	assert.NoError(t, c.Register(&ServeRetryErr.S1ServeErr{})) // should produce an error during the Serve
	assert.NoError(t, c.Init())

	res, err := c.Serve()
	assert.NoError(t, err)

	// we can't be sure, what node will be processed first
	ord := [2]string{"ServeRetryErr.S1ServeErr", "ServeRetryErr.S2ServeErr"}

	count := 0

	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		for r := range res {
			assert.Error(t, r.Error)
			if errors.Is(errors.Serve, r.Error) {
				assert.NoError(t, c.Stop())
				wg.Done()
				return
			}
			if r.VertexID == ord[0] || r.VertexID == ord[1] {
				count++
				if count == 100 {
					assert.NoError(t, c.Stop())
					wg.Done()
					return
				}
			} else {
				assert.Fail(t, "vertex should be in the ord slice")
				wg.Done()
				return
			}
		}
	}()

	wg.Wait()
}

func TestEndure_Serve_Retry_100_With_Random_Err(t *testing.T) {
	c, err := endure.NewContainer(endure.DebugLevel, nil, endure.RetryOnFail(true))
	assert.NoError(t, err)

	assert.NoError(t, c.Register(&ServeRetryErr.S4{}))
	assert.NoError(t, c.Register(&ServeRetryErr.S2{}))
	assert.NoError(t, c.Register(&ServeRetryErr.S2ServeErr{})) // Random error here
	assert.NoError(t, c.Register(&ServeRetryErr.S3Init{}))     // Random Init error here
	assert.NoError(t, c.Register(&ServeRetryErr.S5{}))
	assert.NoError(t, c.Register(&ServeRetryErr.S1ServeErr{})) // should produce an error during the Serve
	assert.NoError(t, c.Init())

	res, err := c.Serve()
	assert.NoError(t, err)

	// we can't be sure, what node will be processed first
	ord := [2]string{"ServeRetryErr.S1ServeErr", "ServeRetryErr.S2ServeErr"}

	count := 0

	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		for r := range res {
			assert.Error(t, r.Error)
			if errors.Is(errors.Serve, r.Error) {
				assert.NoError(t, c.Stop())
				wg.Done()
				return
			}
			if r.VertexID == ord[0] || r.VertexID == ord[1] {
				count++
				if count == 100 {
					assert.NoError(t, c.Stop())
					wg.Done()
					return
				}
			} else {
				assert.Fail(t, "vertex should be in the ord slice")
			}
		}
	}()

	wg.Wait()
}

func TestEndure_NoRegisterInvoke(t *testing.T) {
	c, err := endure.NewContainer(endure.DebugLevel, nil, endure.RetryOnFail(true))
	assert.NoError(t, err)

	assert.Error(t, c.Init())

	_, _ = c.Serve()
	assert.NoError(t, err)

	assert.NoError(t, c.Stop())
}

func TestEndure_CollectorFuncReturnError(t *testing.T) {
	c, err := endure.NewContainer(endure.DebugLevel, nil, endure.RetryOnFail(true))
	assert.NoError(t, err)

	assert.NoError(t, c.Register(&CollectorFuncReturn.FooDep{}))
	assert.NoError(t, c.Register(&CollectorFuncReturn.FooDep2{}))
	assert.Error(t, c.Init())

	_, _ = c.Serve()
	assert.NoError(t, err)

	assert.NoError(t, c.Stop())
}

func TestEndure_ForceExit(t *testing.T) {
	c, err := endure.NewContainer(endure.DebugLevel, nil, endure.RetryOnFail(false)) // stop timeout 10 seconds
	assert.NoError(t, err)

	assert.NoError(t, c.Register(&mixed.Foo{})) // sleep for 15 seconds
	assert.NoError(t, c.Init())

	_, err = c.Serve()
	assert.NoError(t, err)

	assert.Error(t, c.Stop()) // shutdown: timeout exceed, some vertices are not stopped and can cause memory leak
}
