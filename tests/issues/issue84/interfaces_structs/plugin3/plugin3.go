package plugin3

import (
	"context"

	"github.com/roadrunner-server/errors"
)

type Plugin3 struct{}

type Fooer interface {
	Foo() string
}

func (p *Plugin3) Init() error {
	return errors.E(errors.Disabled)
}

func (p *Plugin3) Serve() chan error {
	errCh := make(chan error, 1)
	return errCh
}

func (p *Plugin3) Stop(context.Context) error {
	return nil
}

func (p *Plugin3) Foo() string {
	return "foo"
}
