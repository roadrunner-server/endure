package CyclicDeps

import (
	"context"
)

type Plugin2 struct {
}

func (p2 *Plugin2) Init(p3 *Plugin3) error {
	return nil
}

func (p2 *Plugin2) Serve() chan error {
	errCh := make(chan error, 1)
	return errCh
}

func (p2 *Plugin2) Stop(context.Context) error {
	return nil
}
