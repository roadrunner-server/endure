package plugin4

import (
	"errors"
	"time"
)

var number3 int = 0

type Plugin4 struct {
}

func (f *Plugin4) Init() error {
	return nil
}

func (f *Plugin4) Configure() error {
	return nil
}

func (f *Plugin4) Close() error {
	return nil
}

func (f *Plugin4) Serve() chan error {
	errCh := make(chan error, 1)
	if number3 == 0 {
		number3++
		go func() {
			time.Sleep(time.Second * 3)
			errCh <- errors.New("test error3")
		}()
	} else {
		return nil
	}
	return errCh
}

func (f *Plugin4) Stop() error {
	return nil
}
