package stress

import (
	"sync"
	"testing"
	"time"

	"github.com/spiral/endure"
	"github.com/spiral/endure/tests/stress/depender_func_return"
	"github.com/spiral/endure/tests/stress/init_err"
	"github.com/spiral/endure/tests/stress/serve_err"
	"github.com/spiral/endure/tests/stress/serve_retry_err"
	"github.com/stretchr/testify/assert"
)

func TestEndure_Init_Err(t *testing.T) {
	c, err := endure.NewContainer(endure.DebugLevel, endure.RetryOnFail(false))
	assert.NoError(t, err)

	assert.NoError(t, c.Register(&init_err.S1Err{}))
	assert.NoError(t, c.Register(&init_err.S2Err{})) // should produce an error during the Init
	assert.Error(t, c.Init())
}

func TestEndure_Serve_Err(t *testing.T) {
	c, err := endure.NewContainer(endure.DebugLevel, endure.RetryOnFail(false))
	assert.NoError(t, err)

	assert.NoError(t, c.Register(&serve_err.S4ServeError{})) // should produce an error during the Serve
	assert.NoError(t, c.Register(&serve_err.S2{}))
	assert.NoError(t, c.Register(&serve_err.S3ServeError{}))
	assert.NoError(t, c.Register(&serve_err.S5{}))
	assert.NoError(t, c.Register(&serve_err.S1ServeErr{}))
	err = c.Init()
	if err != nil {
		t.Fatal(err)
	}

	res, err := c.Serve()
	if err != nil {
		t.Fatal(err)
	}

	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		for r := range res { // <--- Error is HERE
			assert.Equal(t, "serve_err.S4ServeError", r.VertexID)
			assert.Error(t, r.Error.Err)
			assert.NoError(t, c.Stop())
			time.Sleep(time.Second * 3)
			wg.Done()
			return
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
*/
func TestEndure_Serve_Retry_Err(t *testing.T) {
	c, err := endure.NewContainer(endure.DebugLevel, endure.RetryOnFail(true))
	assert.NoError(t, err)

	assert.NoError(t, c.Register(&serve_retry_err.S4{}))
	assert.NoError(t, c.Register(&serve_retry_err.S2{}))
	assert.NoError(t, c.Register(&serve_retry_err.S2ServeErr{})) // Random error here
	assert.NoError(t, c.Register(&serve_retry_err.S3{}))
	assert.NoError(t, c.Register(&serve_retry_err.S5{}))
	assert.NoError(t, c.Register(&serve_retry_err.S1ServeErr{})) // should produce an error during the Serve
	assert.NoError(t, c.Init())

	res, err := c.Serve()
	assert.NoError(t, err)

	// we can't be sure, what node will be processed first
	ord := [2]string{"serve_retry_err.S1ServeErr", "serve_retry_err.S2ServeErr"}

	count := 0

	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		for r := range res {
			assert.Error(t, r.Error.Err)
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
	c, err := endure.NewContainer(endure.DebugLevel, endure.RetryOnFail(true))
	assert.NoError(t, err)

	assert.NoError(t, c.Register(&serve_retry_err.S4{}))
	assert.NoError(t, c.Register(&serve_retry_err.S2{}))
	assert.NoError(t, c.Register(&serve_retry_err.S2ServeErr{})) // Random error here
	assert.NoError(t, c.Register(&serve_retry_err.S3{}))
	assert.NoError(t, c.Register(&serve_retry_err.S5{}))
	assert.NoError(t, c.Register(&serve_retry_err.S1ServeErr{})) // should produce an error during the Serve
	assert.NoError(t, c.Init())

	res, err := c.Serve()
	assert.NoError(t, err)

	// we can't be sure, what node will be processed first
	ord := [2]string{"serve_retry_err.S1ServeErr", "serve_retry_err.S2ServeErr"}

	count := 0

	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		for r := range res {
			assert.Error(t, r.Error.Err)
			if r.Error.Code >= 500 {
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
	c, err := endure.NewContainer(endure.DebugLevel, endure.RetryOnFail(true))
	assert.NoError(t, err)

	assert.NoError(t, c.Register(&serve_retry_err.S4{}))
	assert.NoError(t, c.Register(&serve_retry_err.S2{}))
	assert.NoError(t, c.Register(&serve_retry_err.S2ServeErr{})) // Random error here
	assert.NoError(t, c.Register(&serve_retry_err.S3Init{}))     // Random error here
	assert.NoError(t, c.Register(&serve_retry_err.S5{}))
	assert.NoError(t, c.Register(&serve_retry_err.S1ServeErr{})) // should produce an error during the Serve
	assert.NoError(t, c.Init())

	res, err := c.Serve()
	assert.NoError(t, err)

	// we can't be sure, what node will be processed first
	ord := [2]string{"serve_retry_err.S1ServeErr", "serve_retry_err.S2ServeErr"}

	count := 0

	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		for r := range res {
			assert.Error(t, r.Error.Err)
			if r.Error.Code == 501 {
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
	c, err := endure.NewContainer(endure.DebugLevel, endure.RetryOnFail(true))
	assert.NoError(t, err)

	assert.Error(t, c.Init())

	_, _ = c.Serve()
	assert.NoError(t, err)

	assert.NoError(t, c.Stop())
}

func TestEndure_DependerFuncReturnError(t *testing.T) {
	c, err := endure.NewContainer(endure.DebugLevel, endure.RetryOnFail(true))
	assert.NoError(t, err)

	assert.NoError(t, c.Register(&depender_func_return.FooDep{}))
	assert.NoError(t, c.Register(&depender_func_return.FooDep2{}))
	assert.Error(t, c.Init())

	_, _ = c.Serve()
	assert.NoError(t, err)

	assert.NoError(t, c.Stop())
}
