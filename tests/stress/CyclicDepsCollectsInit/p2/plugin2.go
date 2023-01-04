package p2

import (
	"context"
)

type Foo interface {
	Foo() string
}

type Plugin2 struct {
}

func (p2 *Plugin2) Init(p1 Foo) error {
	_ = p1
	return nil
}

func (p2 *Plugin2) Serve() chan error {
	errCh := make(chan error, 1)
	return errCh
}

func (p2 *Plugin2) Stop(context.Context) error {
	return nil
}

func (p2 *Plugin2) Bar() string {
	return "bar"
}
