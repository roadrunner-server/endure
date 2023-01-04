package plugin2

import (
	"context"
	"time"

	"github.com/roadrunner-server/errors"
)

type Plugin2 struct {
}

type I3 interface {
	SomeDepMethod()
}

func (p *Plugin2) Init(I3) error {
	return nil
}

func (p *Plugin2) Serve() chan error {
	errCh := make(chan error, 1)
	time.Sleep(time.Second * 5)
	errCh <- errors.Str("test error from Plugin2")
	return errCh
}

func (p *Plugin2) Stop(context.Context) error {
	return nil
}

func (p *Plugin2) SomeP2DepMethod() {}
