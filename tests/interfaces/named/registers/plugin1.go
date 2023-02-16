package registers

import (
	"context"
)

type Plugin1 struct {
}

func (f *Plugin1) Init(db IDB) error {
	println(db.Name())
	return nil
}

func (f *Plugin1) Serve() chan error {
	errCh := make(chan error)
	return errCh
}

func (f *Plugin1) Stop(context.Context) error {
	return nil
}

func (f *Plugin1) Name() string {
	return "My name is Plugin1, friend!"
}
