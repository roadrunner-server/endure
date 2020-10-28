package plugin2

import "errors"

type Plugin2 struct {
}

func (f *Plugin2) Init() error {
	return errors.New("test backoff error")
}

func (f *Plugin2) Serve() chan error {
	errCh := make(chan error, 1)

	return errCh
}

func (f *Plugin2) Stop() error {
	return nil
}
