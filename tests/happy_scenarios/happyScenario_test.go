package happy_scenarios

import (
	"testing"
	"time"

	"github.com/roadrunner-server/endure/v2"
	"github.com/roadrunner-server/endure/v2/tests/happy_scenarios/plugin1"
	"github.com/roadrunner-server/endure/v2/tests/happy_scenarios/plugin2"
	"github.com/roadrunner-server/endure/v2/tests/happy_scenarios/plugin3"
	"github.com/roadrunner-server/endure/v2/tests/happy_scenarios/plugin4"
	"github.com/roadrunner-server/endure/v2/tests/happy_scenarios/plugin5"
	"github.com/roadrunner-server/endure/v2/tests/happy_scenarios/plugin6"
	"github.com/roadrunner-server/endure/v2/tests/happy_scenarios/plugin7"
	primitive "github.com/roadrunner-server/endure/v2/tests/happy_scenarios/plugin8"
	plugin12 "github.com/roadrunner-server/endure/v2/tests/happy_scenarios/provided_value_but_need_pointer/plugin1"
	plugin22 "github.com/roadrunner-server/endure/v2/tests/happy_scenarios/provided_value_but_need_pointer/plugin2"
	"github.com/stretchr/testify/assert"
	"golang.org/x/exp/slog"
)

func TestEndure_DifferentLogLevels(t *testing.T) {
	testLog(t, slog.LevelDebug)
	testLog(t, slog.LevelError)
	testLog(t, slog.LevelInfo)
	testLog(t, slog.LevelWarn)
}

func testLog(t *testing.T, level slog.Leveler) {
	c := endure.New(level, endure.Visualize())

	assert.NoError(t, c.Register(&plugin4.S4{}))
	assert.NoError(t, c.Register(&plugin2.S2{}))
	assert.NoError(t, c.Register(&plugin3.S3{}))
	assert.NoError(t, c.Register(&plugin1.S1{}))
	assert.NoError(t, c.Register(&plugin5.S5{}))
	assert.NoError(t, c.Register(&plugin6.S6Interface{}))
	assert.NoError(t, c.Init())

	res, err := c.Serve()
	assert.NoError(t, err)

	go func() {
		for r := range res {
			if r.Error != nil {
				assert.NoError(t, r.Error)
				return
			}
		}
	}()

	time.Sleep(time.Second * 2)
	assert.NoError(t, c.Stop())
}

func TestEndure_Init_OK(t *testing.T) {
	c := endure.New(slog.LevelDebug)

	assert.NoError(t, c.Register(&plugin4.S4{}))
	assert.NoError(t, c.Register(&plugin2.S2{}))
	assert.NoError(t, c.Register(&plugin3.S3{}))
	assert.NoError(t, c.Register(&plugin1.S1{}))
	assert.NoError(t, c.Register(&plugin5.S5{}))
	assert.NoError(t, c.Register(&plugin6.S6Interface{}))

	assert.NoError(t, c.Init())

	res, err := c.Serve()
	assert.NoError(t, err)
	go func() {
		for r := range res {
			if r.Error != nil {
				assert.NoError(t, r.Error)
				return
			}
		}
	}()

	time.Sleep(time.Second * 2)
	assert.NoError(t, c.Stop())
}

func TestEndure_DoubleInitDoubleServe_OK(t *testing.T) {
	c := endure.New(slog.LevelDebug)

	assert.NoError(t, c.Register(&plugin4.S4{}))
	assert.NoError(t, c.Register(&plugin2.S2{}))
	assert.NoError(t, c.Register(&plugin3.S3{}))
	assert.NoError(t, c.Register(&plugin1.S1{}))
	assert.NoError(t, c.Register(&plugin5.S5{}))
	assert.NoError(t, c.Register(&plugin6.S6Interface{}))

	assert.NoError(t, c.Init())

	_, err := c.Serve()
	assert.NoError(t, err)
	res, err := c.Serve()
	assert.NoError(t, err)
	go func() {
		for r := range res {
			if r.Error != nil {
				assert.NoError(t, r.Error)
				return
			}
		}
	}()

	time.Sleep(time.Second * 2)
	assert.NoError(t, c.Stop())
}

func TestEndure_Init_1_Element(t *testing.T) {
	c := endure.New(slog.LevelDebug)

	assert.NoError(t, c.Register(&plugin7.Plugin7{}))
	assert.NoError(t, c.Init())

	res, err := c.Serve()
	assert.NoError(t, err)

	go func() {
		for r := range res {
			if r.Error != nil {
				assert.NoError(t, r.Error)
				return
			}
		}
	}()

	time.Sleep(time.Second * 2)

	assert.NoError(t, c.Stop())
}

func TestEndure_ProvidedValueButNeedPointer(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			println("test should panic")
		}
	}()
	c := endure.New(slog.LevelDebug)

	assert.NoError(t, c.Register(&plugin12.Plugin1{}))
	assert.NoError(t, c.Register(&plugin22.Plugin2{}))
	assert.Error(t, c.Init())

	res, err := c.Serve()
	assert.NoError(t, err)

	go func() {
		for r := range res {
			if r.Error != nil {
				assert.NoError(t, r.Error)
				return
			}
		}
	}()

	time.Sleep(time.Second * 2)

	assert.NoError(t, c.Stop())
}

func TestEndure_PrimitiveTypes(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			println("test should panic")
		}
	}()
	c := endure.New(slog.LevelDebug)

	assert.NoError(t, c.Register(&primitive.Plugin8{}))
	assert.Error(t, c.Init())

	_, err := c.Serve()
	assert.NoError(t, err)

	assert.NoError(t, c.Stop())
}

func TestEndure_VisualizeFile(t *testing.T) {
	// digraph endure {\n\trankdir=TB;\n\tgraph [compound=true];\n\t\"*plugin6.S6Interface\" -> \"*plugin4.S4\";\n\t\"*plugin5.S5\" -> \"*plugin4.S4\";\n\t\"*plugin4.S4\" -> \"*plugin2.S2\";\n\t\"*plugin4.S4\" -> \"*plugin3.S3\";\n\t\"*plugin4.S4\" -> \"*plugin1.S1\";\n\t\"*plugin2.S2\" -> \"*plugin3.S3\";\n\t\"*plugin2.S2\" -> \"*plugin1.S1\";\n}\n\t
	// digraph endure {\n\trankdir=TB;\n\tgraph [compound=true];\n\t\"*plugin6.S6Interface\" -> \"*plugin4.S4\";\n\t\"*plugin5.S5\" -> \"*plugin4.S4\";\n\t\"*plugin4.S4\" -> \"*plugin1.S1\";\n\t\"*plugin4.S4\" -> \"*plugin2.S2\";\n\t\"*plugin4.S4\" -> \"*plugin3.S3\";\n\t\"*plugin2.S2\" -> \"*plugin1.S1\";\n\t\"*plugin2.S2\" -> \"*plugin3.S3\";\n}\n\n

	c := endure.New(slog.LevelDebug, endure.Visualize())

	assert.NoError(t, c.Register(&plugin4.S4{}))
	assert.NoError(t, c.Register(&plugin2.S2{}))
	assert.NoError(t, c.Register(&plugin3.S3{}))
	assert.NoError(t, c.Register(&plugin1.S1{}))
	assert.NoError(t, c.Register(&plugin5.S5{}))
	assert.NoError(t, c.Register(&plugin6.S6Interface{}))

	assert.NoError(t, c.Init())
	_, _ = c.Serve()
	assert.NoError(t, c.Stop())
}

func TestEndure_VisualizeStdOut(t *testing.T) {
	c := endure.New(slog.LevelDebug)

	assert.NoError(t, c.Register(&plugin4.S4{}))
	assert.NoError(t, c.Register(&plugin2.S2{}))
	assert.NoError(t, c.Register(&plugin3.S3{}))
	assert.NoError(t, c.Register(&plugin1.S1{}))
	assert.NoError(t, c.Register(&plugin5.S5{}))
	assert.NoError(t, c.Register(&plugin6.S6Interface{}))

	assert.NoError(t, c.Init())
	_, _ = c.Serve()
}
