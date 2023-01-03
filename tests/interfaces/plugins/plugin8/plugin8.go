package plugin8

import (
	"context"

	"github.com/roadrunner-server/endure/v2/dep"
	"github.com/roadrunner-server/errors"
)

type SomeInterface interface {
	Boom()
}

type Named interface {
	// Name return user friendly name of the plugin
	Name() string
}

type Plugin8 struct {
	collectedDeps []any
}

// No deps
func (s *Plugin8) Init() error {
	s.collectedDeps = make([]any, 0, 6)
	return nil
}

func (s *Plugin8) Serve() chan error {
	errCh := make(chan error, 1)
	if len(s.collectedDeps) != 4 {
		errCh <- errors.E("not enough deps collected")
	}
	return errCh
}

func (s *Plugin8) Name() string {
	return "plugin8"
}

func (s *Plugin8) Collects() []*dep.In {
	return []*dep.In{
		dep.Fits(func(p any) {
			s.collectedDeps = append(s.collectedDeps, p)
		}, (*SomeInterface)(nil)),
		dep.Fits(func(p any) {
			s.collectedDeps = append(s.collectedDeps, p)
		}, (*Named)(nil)),
	}
}

func (s *Plugin8) Stop(context.Context) error {
	return nil
}
