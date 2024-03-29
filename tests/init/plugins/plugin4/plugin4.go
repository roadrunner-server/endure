package plugin4

import (
	"context"
	"errors"
	"time"
)

var number3 int = 0

type Plugin4 struct {
}

func (f *Plugin4) Init() error {
	return nil
}

func (f *Plugin4) Serve() chan error {
	errCh := make(chan error, 1)
	if number3 == 0 {
		number3++
		go func() {
			time.Sleep(time.Second * 3)
			errCh <- errors.New("test plugin4 error")
		}()
	} else {
		return errCh
	}
	return errCh
}

func (f *Plugin4) Stop(context.Context) error {
	return nil
}
