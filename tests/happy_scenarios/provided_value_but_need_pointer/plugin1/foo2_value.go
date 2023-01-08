package plugin1

import (
	"context"
)

type DBVp2 interface {
	DBV2()
}

type Plugin1 struct {
}

func (s2 *Plugin1) Init(DBVp2) error {
	return nil
}

func (s2 *Plugin1) Serve() chan error {
	errCh := make(chan error, 1)
	return errCh
}

func (s2 *Plugin1) Stop(context.Context) error {
	return nil
}
