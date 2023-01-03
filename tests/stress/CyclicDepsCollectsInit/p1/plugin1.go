package p1

import (
	"context"
)

type Bar interface {
	Bar() string
}

type Plugin1 struct {
}

func (p1 *Plugin1) Init(bar Bar) error {
	_ = bar
	return nil
}

func (p1 *Plugin1) Serve() chan error {
	errCh := make(chan error, 1)
	return errCh
}

func (p1 *Plugin1) Stop(context.Context) error {
	return nil
}

func (p1 *Plugin1) Foo() string {
	return "foo"
}
