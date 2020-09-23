package plugin2

import "errors"

type Foo struct {
}

func (f *Foo) Init() error {
	return errors.New("test backoff error")
}

func (f *Foo) Serve() chan error {
	errCh := make(chan error, 1)

	return errCh
}

func (f *Foo) Stop() error {
	return nil
}
