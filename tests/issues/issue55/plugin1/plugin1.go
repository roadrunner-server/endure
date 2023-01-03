package plugin1

import (
	"context"
)

type Plugin1 struct {
}

type I2 interface {
	SomeP2DepMethod()
}

func (p *Plugin1) Init(I2) error {
	return nil
}

func (p *Plugin1) Serve() chan error {
	errCh := make(chan error, 1)
	return errCh
}

func (p *Plugin1) Stop(context.Context) error {
	return nil
}
