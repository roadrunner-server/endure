package test3

import (
	"log/slog"
	"testing"

	"github.com/roadrunner-server/endure/v2"
	"github.com/roadrunner-server/endure/v2/tests/general/test3/p1"
	"github.com/roadrunner-server/endure/v2/tests/general/test3/p2"
	"github.com/stretchr/testify/assert"
)

func Test1(t *testing.T) {
	end := endure.New(slog.LevelDebug)

	err := end.Register(&p1.Plugin{})
	assert.NoError(t, err)

	err = end.Register(&p2.Plugin{})
	assert.NoError(t, err)

	err = end.Init()
	assert.NoError(t, err)

	_, err = end.Serve()
	assert.NoError(t, err)

	assert.NoError(t, end.Stop())
}
