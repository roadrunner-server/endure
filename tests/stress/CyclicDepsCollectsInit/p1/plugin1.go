package p1

import (
	"github.com/roadrunner-server/endure/tests/stress/CyclicDepsCollectsInit/api/p2"
)

type Plugin1 struct {
}

func (p1 *Plugin1) Init(bar p2.Bar) error {
	_ = bar
	return nil
}

func (p1 *Plugin1) Serve() chan error {
	errCh := make(chan error, 1)
	return errCh
}

func (p1 *Plugin1) Stop() error {
	return nil
}

func (p1 *Plugin1) Foo() string {
	return "foo"
}
