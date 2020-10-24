package plugin1

import (
	"errors"
)

var number int = 0

type Plugin1 struct {
}

func (f *Plugin1) Init() error {
	number += 1
	return nil
}

func (f *Plugin1) Serve() chan error {
	errCh := make(chan error, 1)
	go func() {
		errCh <- errors.New("test error in serve")
	}()
	return errCh
}

func (f *Plugin1) Stop() error {
	return nil
}
