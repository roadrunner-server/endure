package randominterface

import (
	"context"
)

type Plugin1 struct {
}

type SuperInterface interface {
	Super() string
}

func (f *Plugin1) Init(db SuperInterface) error {
	println(db.Super())
	return nil
}

func (f *Plugin1) Serve() chan error {
	errCh := make(chan error)
	return errCh
}

func (f *Plugin1) Stop(context.Context) error {
	return nil
}

func (f *Plugin1) Super() string {
	return "SUPER -> "
}
