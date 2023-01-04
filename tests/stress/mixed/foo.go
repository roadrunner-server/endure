package mixed

import (
	"context"
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

func (f *Foo) Stop(ctx context.Context) error {
	fin := make(chan struct{}, 1)
	go func() {
		time.Sleep(time.Second * 15)
		fin <- struct{}{}
	}()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-fin:
		return nil
	}
}
