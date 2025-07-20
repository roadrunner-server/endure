package CyclicDeps

import (
	"context"
)

type Plugin3 struct {
}

type Stringer1 interface {
	String1() string
}

func (p3 *Plugin3) Init(p1 Stringer1) error {
	return nil
}

func (p3 *Plugin3) Serve() chan error {
	errCh := make(chan error, 1)
	return errCh
}

func (p3 *Plugin3) Stop(context.Context) error {
	return nil
}

func (p3 *Plugin3) String3() string {
	return "Plugin3"
}
