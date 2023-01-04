package p1

import (
	"context"

	"github.com/roadrunner-server/endure/v2/dep"
)

type Bar interface {
	Bar() string
}

type Plugin1 struct {
}

func (p1 *Plugin1) Init() error {
	return nil
}

func (p1 *Plugin1) Serve() chan error {
	errCh := make(chan error, 1)
	return errCh
}

func (p1 *Plugin1) Stop(context.Context) error {
	return nil
}

func (p1 *Plugin1) Collects() []*dep.In {
	return []*dep.In{
		dep.Fits(func(any) {

		}, (*Bar)(nil)),
	}
}

func (p1 *Plugin1) Foo() string {
	return "foo"
}
