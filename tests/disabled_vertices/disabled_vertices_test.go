package disabled_vertices

import (
	"testing"

	"github.com/roadrunner-server/endure/v2"
	"github.com/roadrunner-server/endure/v2/tests/disabled_vertices/plugin1"
	"github.com/roadrunner-server/endure/v2/tests/disabled_vertices/plugin2"
	"github.com/roadrunner-server/endure/v2/tests/disabled_vertices/plugin3"
	"github.com/roadrunner-server/endure/v2/tests/disabled_vertices/plugin4"
	"github.com/roadrunner-server/endure/v2/tests/disabled_vertices/plugin5"
	"github.com/roadrunner-server/endure/v2/tests/disabled_vertices/plugin6"
	"github.com/roadrunner-server/endure/v2/tests/disabled_vertices/plugin7"
	"github.com/roadrunner-server/endure/v2/tests/disabled_vertices/plugin8"
	"github.com/roadrunner-server/endure/v2/tests/disabled_vertices/plugin9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/exp/slog"
)

func TestVertexDisabled(t *testing.T) {
	cont := endure.New(slog.LevelDebug)
	err := cont.Register(&plugin1.Plugin1{}) // disabled
	if err != nil {
		t.Fatal(err)
	}

	err = cont.Register(&plugin2.Plugin2{}) // depends via init
	if err != nil {
		t.Fatal(err)
	}

	err = cont.Init()
	assert.Error(t, err)
	_ = cont.Stop()
}

func TestDisabledViaInterface(t *testing.T) {
	cont := endure.New(slog.LevelDebug)

	err := cont.Register(&plugin3.Plugin3{}) // disabled
	if err != nil {
		t.Fatal(err)
	}

	err = cont.Register(&plugin4.Plugin4{}) // depends via init
	if err != nil {
		t.Fatal(err)
	}

	err = cont.Register(&plugin5.Plugin5{}) // depends via init
	if err != nil {
		t.Fatal(err)
	}

	err = cont.Init()
	assert.NoError(t, err)
}

func TestDisabledRoot(t *testing.T) {
	cont := endure.New(slog.LevelDebug)

	err := cont.Register(&plugin6.Plugin6{}) // Root
	if err != nil {
		t.Fatal(err)
	}

	err = cont.Register(&plugin7.Plugin7{}) // should be disabled
	if err != nil {
		t.Fatal(err)
	}

	err = cont.Register(&plugin8.Plugin8{}) // should be disabled
	if err != nil {
		t.Fatal(err)
	}

	err = cont.Register(&plugin9.Plugin9{}) // should be disabled
	if err != nil {
		t.Fatal(err)
	}

	err = cont.Init()
	assert.Error(t, err)
}

func TestOneSurvived(t *testing.T) {
	cont := endure.New(slog.LevelDebug)

	err := cont.RegisterAll(
		&plugin6.Plugin6{},
		&plugin7.Plugin7{},
		&plugin8.Plugin8{},
		&plugin9.Plugin9{},
		&plugin5.Plugin5{},
	)

	require.NoError(t, err)
	err = cont.Init()
	assert.NoError(t, err)

	_, err = cont.Serve()
	require.NoError(t, err)
}
