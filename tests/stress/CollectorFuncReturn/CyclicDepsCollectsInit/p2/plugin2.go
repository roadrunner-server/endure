package p2

import (
	"github.com/spiral/endure/tests/stress/CyclicDepsCollects/api/p1"
)

type Plugin2 struct {
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

func (p2 *Plugin2) Collects() []interface{} {
	return []interface{}{
		p2.GetP1,
	}
}

func (p2 *Plugin2) Bar() string {
	return "bar"
}

func (p2 *Plugin2) GetP1(p1 p1.Foo) {
	_ = p1
}
