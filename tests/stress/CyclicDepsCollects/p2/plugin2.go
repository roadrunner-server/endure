package p2

import (
	"github.com/roadrunner-server/endure/v2/dep"
)

type Plugin2 struct {
}

type Foo interface {
	Foo() string
}

func (p2 *Plugin2) Init() error {
	return nil
}

func (p2 *Plugin2) Serve() chan error {
	errCh := make(chan error, 1)
	return errCh
}

func (p2 *Plugin2) Stop() error {
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
