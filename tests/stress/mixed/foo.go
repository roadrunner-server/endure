package mixed

import (
	"time"
)

type Foo struct {
}

func (f *Foo) Init() error {
	return nil
}

func (f *Foo) Serve() chan error {
	errCh := make(chan error, 1)
	return errCh
}

func (f *Foo) Stop() error {
	time.Sleep(time.Second * 15)
	return nil
}
