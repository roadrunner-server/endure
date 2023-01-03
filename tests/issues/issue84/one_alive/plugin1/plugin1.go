package plugin1

import (
	"context"
)

type Plugin1 struct {
}

type Fooer interface {
	Foo() string
}

func (p *Plugin1) Init(p3 Fooer) error {
	println(p3.Foo())
	return nil
}

func (p *Plugin1) Serve() chan error {
	errCh := make(chan error, 1)
	return errCh
}

func (p *Plugin1) Stop(context.Context) error {
	return nil
}
