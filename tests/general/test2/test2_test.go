package test1

import (
	"testing"

	"github.com/roadrunner-server/endure/v2"
	"github.com/roadrunner-server/endure/v2/tests/general/test2/p1"
	"github.com/roadrunner-server/endure/v2/tests/general/test2/p2"
	"github.com/roadrunner-server/endure/v2/tests/general/test2/p3"
	"github.com/roadrunner-server/endure/v2/tests/general/test2/p4"
	"github.com/roadrunner-server/endure/v2/tests/general/test2/p5"
	"github.com/roadrunner-server/endure/v2/tests/general/test2/p6"
	"github.com/stretchr/testify/assert"
	"golang.org/x/exp/slog"
)

func Test1(t *testing.T) {
	end := endure.New(slog.LevelDebug)

	err := end.Register(&p1.Plugin{})
	assert.NoError(t, err)

	err = end.Register(&p2.Plugin{})
	assert.NoError(t, err)

	err = end.Register(&p3.Plugin{})
	assert.NoError(t, err)

	err = end.Register(&p4.Plugin{})
	assert.NoError(t, err)

	err = end.Register(&p5.Plugin{})
	assert.NoError(t, err)

	err = end.Register(&p6.Plugin{})
	assert.NoError(t, err)

	err = end.Init()
	assert.NoError(t, err)

	_, err = end.Serve()
	assert.NoError(t, err)

	assert.NoError(t, end.Stop())
}

func Test2(t *testing.T) {
	end := endure.New(slog.LevelDebug)

	err := end.Register(&p3.Plugin{})
	assert.NoError(t, err)

	err = end.Register(&p4.Plugin{})
	assert.NoError(t, err)

	err = end.Init()
	assert.NoError(t, err)
}
