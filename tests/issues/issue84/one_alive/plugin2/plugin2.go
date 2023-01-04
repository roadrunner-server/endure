package plugin2

import (
	"context"

	"github.com/roadrunner-server/errors"
)

type Plugin2 struct{}

type Fooer interface {
	Foo() string
}

type I3 interface {
	SomeP3DepMethod()
}

func (p *Plugin2) Init(_ Fooer, _ I3) error {
	return errors.E(errors.Disabled)
}

func (p *Plugin2) Serve() chan error {
	errCh := make(chan error, 1)
	return errCh
}

func (p *Plugin2) Stop(context.Context) error {
	return nil
}
