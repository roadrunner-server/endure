package CyclicDeps

import (
	"context"
)

type Plugin1 struct {
}

type Stringer2 interface {
	String2() string
}

func (p1 *Plugin1) Init(p2 Stringer2) error {
	return nil
}

func (p1 *Plugin1) Serve() chan error {
	errCh := make(chan error, 1)
	return errCh
}

func (p1 *Plugin1) Stop(context.Context) error {
	return nil
}

func (p1 *Plugin1) String1() string {
	return "Plugin1"
}
