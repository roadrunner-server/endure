package plugin1

import (
	"context"
)

type Plugin1 struct {
}

type P2Dep interface {
	SomeMethodP2()
}

func (p *Plugin1) Init(P2Dep) error {
	return nil
}

func (p *Plugin1) Serve() chan error {
	errCh := make(chan error, 1)
	return errCh
}

func (p *Plugin1) Stop(context.Context) error {
	return nil
}
