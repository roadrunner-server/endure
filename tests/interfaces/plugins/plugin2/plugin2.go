package plugin2

import (
	"context"

	"github.com/roadrunner-server/endure/v2/dep"
)

type FooWriter interface {
	Fooo() // just stupid name
}

type Plugin2 struct {
}

func (s *Plugin2) Fooo() {
	println("just FooWriter interface invoke")
}

// No deps
func (s *Plugin2) Init() error {
	return nil
}

func (s *Plugin2) Serve() chan error {
	errCh := make(chan error, 1)
	return errCh
}

func (s *Plugin2) Stop(context.Context) error {
	return nil
}

func (s *Plugin2) Provides() []*dep.Out {
	return []*dep.Out{
		dep.OutType((*FooWriter)(nil), s.ProvideInterface),
	}
}

func (s *Plugin2) ProvideInterface() (FooWriter, error) {
	return &Plugin2{}, nil
}
