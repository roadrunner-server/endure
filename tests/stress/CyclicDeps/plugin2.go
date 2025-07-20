package CyclicDeps

import (
	"context"
)

type Plugin2 struct {
}

type Stringer3 interface {
	String3() string
}

func (p2 *Plugin2) Init(p3 Stringer3) error {
	return nil
}

func (p2 *Plugin2) Serve() chan error {
	errCh := make(chan error, 1)
	return errCh
}

func (p2 *Plugin2) Stop(context.Context) error {
	return nil
}

func (p2 *Plugin2) String2() string {
	return "Plugin2"
}
