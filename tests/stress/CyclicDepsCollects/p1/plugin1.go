package p1

import (
	"github.com/spiral/endure/tests/stress/CyclicDepsCollects/api/p2"
)

type Plugin1 struct {
}

func (p1 *Plugin1) Init() error {
	return nil
}

func (p1 *Plugin1) Serve() chan error {
	errCh := make(chan error, 1)
	return errCh
}

func (p1 *Plugin1) Stop() error {
	return nil
}

func (p1 *Plugin1) Collects() []interface{} {
	return []interface{}{
		p1.GetP2,
	}
}

func (p1 *Plugin1) Foo() string {
	return "foo"
}

func (p1 *Plugin1) GetP2(bar p2.Bar) {
	_ = bar
}
