package issue445_test

import (
	"context"
)

type Plugin2 struct {
}

func (p *Plugin2) Init() error {
	return nil
}

func (p *Plugin2) Serve() chan error {
	errCh := make(chan error, 1)
	return errCh
}

func (p *Plugin2) Stop(context.Context) error {
	return nil
}
