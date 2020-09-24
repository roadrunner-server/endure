package issue33

import "errors"

type Plugin1 struct {
}

func (f *Plugin1) Init() error {
	return nil
}

func (f *Plugin1) Serve() chan error {
	errCh := make(chan error, 1)

	return errCh
}

func (f *Plugin1) Stop() error {
	return nil
}

func (f *Plugin1) Provides() []interface{} {
	return []interface{}{
		f, // <- should be a function
	}
}

func (f *Plugin1) AddService(dep2 Plugin2) error {
	return errors.New("test dependers error")
}

type Plugin2 struct {
}

func (f *Plugin2) Init() error {
	return nil
}

func (f *Plugin2) Serve() chan error {
	errCh := make(chan error, 1)

	return errCh
}

func (f *Plugin2) Stop() error {
	return nil
}
