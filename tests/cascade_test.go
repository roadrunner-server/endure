package cascade_test

import (
	"sync"
	"testing"
	"time"

	"github.com/spiral/cascade/tests/backofftimertest"
	"github.com/spiral/cascade/tests/dependers/returnerr"
	"github.com/spiral/cascade/tests/foo5"
	"github.com/spiral/cascade/tests/foo6"
	"github.com/spiral/cascade/tests/foo7"
	"github.com/spiral/cascade/tests/foo8"
	"github.com/spiral/cascade/tests/foo9"
	"github.com/spiral/cascade/tests/primitive"
	"github.com/spiral/cascade/tests/registers/named/randominterface"
	"github.com/spiral/cascade/tests/registers/named/registers"
	"github.com/spiral/cascade/tests/registers/named/registersfail"
	"github.com/stretchr/testify/assert"

	"github.com/spiral/cascade"
	"github.com/spiral/cascade/tests/foo1"
	"github.com/spiral/cascade/tests/foo2"
	"github.com/spiral/cascade/tests/foo3"
	"github.com/spiral/cascade/tests/foo4"
)

func TestCascade_NoRegisterInvoke(t *testing.T) {
	c, err := cascade.NewContainer(cascade.DebugLevel, cascade.RetryOnFail(true))
	assert.NoError(t, err)

	assert.Error(t, c.Init())

	_, _ = c.Serve()
	assert.NoError(t, err)

	assert.NoError(t, c.Stop())
}

func TestCascade_DependerFuncReturnError(t *testing.T) {
	c, err := cascade.NewContainer(cascade.DebugLevel, cascade.RetryOnFail(true))
	assert.NoError(t, err)

	assert.NoError(t, c.Register(&returnerr.FooDep{}))
	assert.NoError(t, c.Register(&returnerr.FooDep2{}))
	assert.Error(t, c.Init())

	_, _ = c.Serve()
	assert.NoError(t, err)

	assert.NoError(t, c.Stop())
}

func TestCascade_BackoffTimers(t *testing.T) {
	c, err := cascade.NewContainer(cascade.DebugLevel, cascade.RetryOnFail(true), cascade.SetBackoffTimes(time.Second, time.Second * 5))
	assert.NoError(t, err)

	assert.NoError(t, c.Register(&backofftimertest.Foo{}))
	assert.Error(t, c.Init())

	_, _ = c.Serve()
	assert.NoError(t, err)

	assert.NoError(t, c.Stop())
}

func TestCascade_PrimitiveTypes(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			println("test should panic")
		}
	}()
	c, err := cascade.NewContainer(cascade.DebugLevel)
	assert.NoError(t, err)

	assert.NoError(t, c.Register(&primitive.Foo{}))
	assert.NoError(t, c.Init())

	_, _ = c.Serve()
	assert.NoError(t, err)

	assert.NoError(t, c.Stop())
}

func TestCascade_Init_OK(t *testing.T) {
	c, err := cascade.NewContainer(cascade.DebugLevel)
	assert.NoError(t, err)

	assert.NoError(t, c.Register(&foo4.S4{}))
	assert.NoError(t, c.Register(&foo2.S2{}))
	assert.NoError(t, c.Register(&foo3.S3{}))
	assert.NoError(t, c.Register(&foo1.S1{}))
	assert.NoError(t, c.Register(&foo5.S5{}))
	assert.NoError(t, c.Register(&foo6.S6Interface{}))
	assert.NoError(t, c.Init())

	res, err := c.Serve()
	assert.NoError(t, err)

	go func() {
		for r := range res {
			if r.Error.Err != nil {
				assert.NoError(t, r.Error.Err)
				return
			}
		}
	}()

	time.Sleep(time.Second * 2)

	assert.NoError(t, c.Stop())
	time.Sleep(time.Second * 1)
}

func TestCascade_Interfaces_OK(t *testing.T) {
	c, err := cascade.NewContainer(cascade.DebugLevel)
	assert.NoError(t, err)

	assert.NoError(t, c.Register(&foo5.S5Interface{}))
	assert.NoError(t, c.Register(&foo6.S6Interface{}))
	err = c.Init()
	if err != nil {
		t.Fatal(err)
	}

	res, err := c.Serve()
	assert.NoError(t, err)

	go func() {
		for r := range res {
			if r.Error.Err != nil {
				assert.NoError(t, r.Error.Err)
				return
			}
		}
	}()

	time.Sleep(time.Second * 2)

	assert.NoError(t, c.Stop())
	time.Sleep(time.Second * 1)
}

func TestCascade_Init_1_Element(t *testing.T) {
	c, err := cascade.NewContainer(cascade.DebugLevel)
	assert.NoError(t, err)

	assert.NoError(t, c.Register(&foo1.S1One{}))
	assert.NoError(t, c.Init())

	res, err := c.Serve()
	assert.NoError(t, err)

	go func() {
		for r := range res {
			if r.Error.Err != nil {
				assert.NoError(t, r.Error.Err)
				return
			}
		}
	}()

	time.Sleep(time.Second * 2)

	assert.NoError(t, c.Stop())
	time.Sleep(time.Second * 1)
}

func TestCascade_ProvidedValueButNeedPointer(t *testing.T) {
	c, err := cascade.NewContainer(cascade.DebugLevel)
	assert.NoError(t, err)

	assert.NoError(t, c.Register(&foo2.S2V{}))
	assert.NoError(t, c.Register(&foo4.S4V{}))
	assert.NoError(t, c.Init())

	res, err := c.Serve()
	assert.NoError(t, err)

	go func() {
		for r := range res {
			if r.Error.Err != nil {
				assert.NoError(t, r.Error.Err)
				return
			}
		}
	}()

	time.Sleep(time.Second * 2)

	assert.NoError(t, c.Stop())
	time.Sleep(time.Second * 1)
}

func TestCascade_Init_Err(t *testing.T) {
	c, err := cascade.NewContainer(cascade.DebugLevel, cascade.RetryOnFail(false))
	assert.NoError(t, err)

	assert.NoError(t, c.Register(&foo1.S1Err{}))
	assert.NoError(t, c.Register(&foo2.S2Err{})) // should produce an error during the Init
	assert.Error(t, c.Init())
}

func TestCascade_Serve_Err(t *testing.T) {
	c, err := cascade.NewContainer(cascade.DebugLevel, cascade.RetryOnFail(false))
	assert.NoError(t, err)

	assert.NoError(t, c.Register(&foo4.S4ServeError{}))
	assert.NoError(t, c.Register(&foo2.S2{}))
	assert.NoError(t, c.Register(&foo3.S3ServeError{}))
	assert.NoError(t, c.Register(&foo5.S5{}))
	assert.NoError(t, c.Register(&foo1.S1ServeErr{})) // should produce an error during the Serve
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
		for r := range res { //<--- Error is HERE
			assert.Equal(t, "foo4.S4ServeError", r.VertexID)
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
func TestCascade_Serve_Retry_Err(t *testing.T) {
	c, err := cascade.NewContainer(cascade.DebugLevel, cascade.RetryOnFail(true))
	assert.NoError(t, err)

	assert.NoError(t, c.Register(&foo4.S4{}))
	assert.NoError(t, c.Register(&foo2.S2{}))
	assert.NoError(t, c.Register(&foo2.S2ServeErr{})) // Random error here
	assert.NoError(t, c.Register(&foo3.S3{}))
	assert.NoError(t, c.Register(&foo5.S5{}))
	assert.NoError(t, c.Register(&foo1.S1ServeErr{})) // should produce an error during the Serve
	assert.NoError(t, c.Init())

	res, err := c.Serve()
	assert.NoError(t, err)

	// we can't be sure, what node will be processed first
	ord := [2]string{"foo1.S1ServeErr", "foo2.S2ServeErr"}

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
func TestCascade_Serve_Retry_100_Err(t *testing.T) {
	c, err := cascade.NewContainer(cascade.DebugLevel, cascade.RetryOnFail(true))
	assert.NoError(t, err)

	assert.NoError(t, c.Register(&foo4.S4{}))
	assert.NoError(t, c.Register(&foo2.S2{}))
	assert.NoError(t, c.Register(&foo2.S2ServeErr{})) // Random error here
	assert.NoError(t, c.Register(&foo3.S3{}))
	assert.NoError(t, c.Register(&foo5.S5{}))
	assert.NoError(t, c.Register(&foo1.S1ServeErr{})) // should produce an error during the Serve
	assert.NoError(t, c.Init())

	res, err := c.Serve()
	assert.NoError(t, err)

	// we can't be sure, what node will be processed first
	ord := [2]string{"foo1.S1ServeErr", "foo2.S2ServeErr"}

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

func TestCascade_Serve_Retry_100_With_Random_Err(t *testing.T) {
	c, err := cascade.NewContainer(cascade.DebugLevel, cascade.RetryOnFail(true))
	assert.NoError(t, err)

	assert.NoError(t, c.Register(&foo4.S4{}))
	assert.NoError(t, c.Register(&foo2.S2{}))
	assert.NoError(t, c.Register(&foo2.S2ServeErr{})) // Random error here
	assert.NoError(t, c.Register(&foo3.S3Init{}))     // Random error here
	assert.NoError(t, c.Register(&foo5.S5{}))
	assert.NoError(t, c.Register(&foo1.S1ServeErr{})) // should produce an error during the Serve
	assert.NoError(t, c.Init())

	res, err := c.Serve()
	assert.NoError(t, err)

	// we can't be sure, what node will be processed first
	ord := [2]string{"foo1.S1ServeErr", "foo2.S2ServeErr"}

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

func TestCascade_InterfacesDepends_Ok(t *testing.T) {
	c, err := cascade.NewContainer(cascade.DebugLevel)
	assert.NoError(t, err)

	assert.NoError(t, c.Register(&foo7.Foo7{}))
	assert.NoError(t, c.Register(&foo8.Foo8{}))
	assert.NoError(t, c.Register(&foo9.Foo9{}))

	assert.NoError(t, c.Init())

	_, err = c.Serve()
	assert.NoError(t, err)

	assert.NoError(t, c.Stop())
}

func TestCascade_NamedProvides_Ok(t *testing.T) {
	c, err := cascade.NewContainer(cascade.DebugLevel)
	assert.NoError(t, err)

	assert.NoError(t, c.Register(&registers.Foo11{}))
	assert.NoError(t, c.Register(&registers.Foo10{}))

	assert.NoError(t, c.Init())

	_, err = c.Serve()
	assert.NoError(t, err)

	assert.NoError(t, c.Stop())
}

func TestCascade_NamedProvides_NotImplement_Ok(t *testing.T) {
	c, err := cascade.NewContainer(cascade.DebugLevel)
	assert.NoError(t, err)

	assert.NoError(t, c.Register(&randominterface.Foo1{}))
	assert.NoError(t, c.Register(&randominterface.Foo{}))

	assert.NoError(t, c.Init())

	_, err = c.Serve()
	assert.NoError(t, err)

	assert.NoError(t, c.Stop())
}

func TestCascade_NamedProvides_WrongType_Fail(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			println("test should panic")
		}
	}()
	c, err := cascade.NewContainer(cascade.DebugLevel)
	assert.NoError(t, err)

	assert.NoError(t, c.Register(&registersfail.Foo1{}))
	assert.NoError(t, c.Register(&registersfail.Foo{}))

	assert.Error(t, c.Init())

	_, err = c.Serve()
	assert.NoError(t, err)

	assert.NoError(t, c.Stop())
}
