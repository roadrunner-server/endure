package plugin1

import (
	"context"
	"errors"
)

type Plugin1 struct {
}

func (f *Plugin1) Init() error {
	return nil
}

func (f *Plugin1) Serve() chan error {
	errCh := make(chan error, 1)
	go func() {
		errCh <- errors.New("test error in serve")
	}()
	return errCh
}

func (f *Plugin1) Stop(context.Context) error {
	return nil
}
