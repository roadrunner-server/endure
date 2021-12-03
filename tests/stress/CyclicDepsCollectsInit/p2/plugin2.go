package p2

import (
	"github.com/spiral/endure/tests/stress/CyclicDepsCollectsInit/api/p1"
)

type Plugin2 struct {
}

func (p2 *Plugin2) Init(p1 p1.Foo) error {
	_ = p1
	return nil
}

func (p2 *Plugin2) Serve() chan error {
	errCh := make(chan error, 1)
	return errCh
}

func (p2 *Plugin2) Stop() error {
	return nil
}

func (p2 *Plugin2) Bar() string {
	return "bar"
}
