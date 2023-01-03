package CyclicDeps

import (
	"context"
)

type Plugin1 struct {
}

func (p1 *Plugin1) Init(p2 *Plugin2) error {
	return nil
}

func (p1 *Plugin1) Serve() chan error {
	errCh := make(chan error, 1)
	return errCh
}

func (p1 *Plugin1) Stop(context.Context) error {
	return nil
}
