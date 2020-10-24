package plugin1

import (
	"github.com/spiral/endure/tests/happy_scenarios/plugin6"
)

type Plugin1 struct {
}

// No deps
func (s *Plugin1) Init(foow plugin6.FooWriter) error {
	foow.Fooo()
	return nil
}

func (s *Plugin1) Configure() error {
	return nil
}

func (s *Plugin1) Serve() chan error {
	errCh := make(chan error, 1)
	return errCh
}

func (s *Plugin1) Close() error {
	return nil
}

func (s *Plugin1) Stop() error {
	return nil
}
