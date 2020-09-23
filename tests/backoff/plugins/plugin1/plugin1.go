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

func (f *Plugin1) Configure() error {
	if number > 1 {
		return errors.New("test error when num > 1")
	}
	return nil
}

func (f *Plugin1) Close() error {
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
