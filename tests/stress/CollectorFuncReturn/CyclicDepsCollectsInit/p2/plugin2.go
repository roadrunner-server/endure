package p2

import (
	"context"

	"github.com/roadrunner-server/endure/v2/dep"
)

type Foo interface {
	Foo() string
}

type Plugin2 struct {
}

func (p2 *Plugin2) Init() error {
	return nil
}

func (p2 *Plugin2) Serve() chan error {
	errCh := make(chan error, 1)
	return errCh
}

func (p2 *Plugin2) Stop(context.Context) error {
	return nil
}

func (p2 *Plugin2) Collects() []*dep.In {
	return []*dep.In{
		dep.Fits(func(any) {

		}, (*Foo)(nil)),
	}
}

func (p2 *Plugin2) Bar() string {
	return "bar"
}
