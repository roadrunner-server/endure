package testcase33

import "errors"

type FooDep struct {
}

func (f *FooDep) Init() error {
	return nil
}

func (f *FooDep) Serve() chan error {
	errCh := make(chan error, 1)

	return errCh
}

func (f *FooDep) Stop() error {
	return nil
}

func (f *FooDep) Provides() []interface{} {
	return []interface{}{
		f, // <- should be a function
	}
}

func (f *FooDep) AddService(dep2 FooDep2) error {
	return errors.New("test dependers error")
}

type FooDep2 struct {
}

func (f *FooDep2) Init() error {
	return nil
}

func (f *FooDep2) Serve() chan error {
	errCh := make(chan error, 1)

	return errCh
}

func (f *FooDep2) Stop() error {
	return nil
}
