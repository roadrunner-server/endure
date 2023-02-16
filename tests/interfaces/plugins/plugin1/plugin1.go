package plugin1

import (
	"context"
)

type FooWriter interface {
	Fooo() // just stupid name
}

type Plugin1 struct {
}

// No deps
func (s *Plugin1) Init(foow FooWriter) error {
	foow.Fooo()
	return nil
}

func (s *Plugin1) Serve() chan error {
	errCh := make(chan error, 1)
	return errCh
}

func (s *Plugin1) Stop(context.Context) error {
	return nil
}
