package plugin3

import (
	"context"
	"errors"
	"time"
)

var number2 int = 0

type Plugin3 struct {
}

func (f *Plugin3) Init() error {
	if number2 > 0 {
		return errors.New("test error when num > 1")
	}
	return nil
}

func (f *Plugin3) Serve() chan error {
	errCh := make(chan error, 1)
	number2 += 1
	go func() {
		time.Sleep(time.Millisecond * 500)
		errCh <- errors.New("test error2")
	}()
	return errCh
}

func (f *Plugin3) Stop(context.Context) error {
	return nil
}
