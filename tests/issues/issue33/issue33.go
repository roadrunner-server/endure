package issue33

import (
	"context"

	"github.com/roadrunner-server/endure/v2/dep"
)

type Plugin1 struct {
}

func (f *Plugin1) Init() error {
	return nil
}

func (f *Plugin1) Serve() chan error {
	errCh := make(chan error, 1)

	return errCh
}

func (f *Plugin1) Stop(context.Context) error {
	return nil
}

func (f *Plugin1) Provides() []*dep.Out {
	return []*dep.Out{}
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

func (f *Plugin2) Stop(context.Context) error {
	return nil
}
