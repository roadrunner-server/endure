package stress

import (
	"testing"
	"time"

	"github.com/roadrunner-server/endure/v2"
	"github.com/roadrunner-server/endure/v2/tests/stress/CyclicDeps"
	"github.com/roadrunner-server/endure/v2/tests/stress/CyclicDepsCollects/p1"
	"github.com/roadrunner-server/endure/v2/tests/stress/CyclicDepsCollects/p2"
	p1Init "github.com/roadrunner-server/endure/v2/tests/stress/CyclicDepsCollectsInit/p1"
	p2Init "github.com/roadrunner-server/endure/v2/tests/stress/CyclicDepsCollectsInit/p2"
	"github.com/roadrunner-server/endure/v2/tests/stress/InitErr"
	"github.com/roadrunner-server/endure/v2/tests/stress/ServeErr"
	"github.com/roadrunner-server/endure/v2/tests/stress/mixed"
	"github.com/stretchr/testify/assert"
	"golang.org/x/exp/slog"
)

func TestEndure_Init_Err(t *testing.T) {
	c := endure.New(slog.LevelDebug)

	assert.NoError(t, c.Register(&InitErr.S1Err{}))
	assert.NoError(t, c.Register(&InitErr.S2Err{})) // should produce an error during the Init
	assert.Error(t, c.Init())
}

func TestEndure_DoubleStop_Err(t *testing.T) {
	c := endure.New(slog.LevelDebug)

	assert.NoError(t, c.Register(&InitErr.S1Err{}))
	assert.NoError(t, c.Register(&InitErr.S2Err{})) // should produce an error during the Init
	assert.Error(t, c.Init())
	assert.Error(t, c.Stop())
	assert.Error(t, c.Stop())
}

func TestEndure_Serve_Err(t *testing.T) {
	c := endure.New(slog.LevelDebug)

	assert.NoError(t, c.Register(&ServeErr.S4ServeError{})) // should produce an error during the Serve
	assert.NoError(t, c.Register(&ServeErr.S2{}))
	assert.NoError(t, c.Register(&ServeErr.S3ServeError{}))
	assert.NoError(t, c.Register(&ServeErr.S5{}))
	assert.NoError(t, c.Register(&ServeErr.S1ServeErr{}))
	err := c.Init()
	if err != nil {
		t.Fatal(err)
	}

	_, err = c.Serve()
	assert.Error(t, err)
}

func TestEndure_NoRegisterInvoke(t *testing.T) {
	c := endure.New(slog.LevelDebug)

	assert.Error(t, c.Init())

	_, err := c.Serve()
	assert.Error(t, err)

	assert.Error(t, c.Stop())
}

func TestEndure_ForceExit(t *testing.T) {
	c := endure.New(slog.LevelDebug, endure.GracefulShutdownTimeout(time.Second)) // stop timeout 10 seconds

	assert.NoError(t, c.Register(&mixed.Foo{})) // sleep for 15 seconds
	assert.NoError(t, c.Init())

	_, err := c.Serve()
	assert.NoError(t, err)

	assert.Error(t, c.Stop()) // shutdown: timeout exceed, some vertices are not stopped and can cause memory leak
}

func TestEndure_CyclicDeps(t *testing.T) {
	c := endure.New(slog.LevelDebug)

	assert.NoError(t, c.RegisterAll(
		&CyclicDeps.Plugin1{},
		&CyclicDeps.Plugin2{},
		&CyclicDeps.Plugin3{},
	))

	assert.Error(t, c.Init())
}

func TestEndure_CyclicDepsCollects(t *testing.T) {
	c := endure.New(slog.LevelDebug)

	assert.NoError(t, c.RegisterAll(
		&p1.Plugin1{},
		&p2.Plugin2{},
	))

	assert.Error(t, c.Init())
}

func TestEndure_CyclicDepsInterfaceInit(t *testing.T) {
	c := endure.New(slog.LevelDebug)

	assert.NoError(t, c.RegisterAll(
		&p1Init.Plugin1{},
		&p2Init.Plugin2{},
	))

	assert.Error(t, c.Init())
}
