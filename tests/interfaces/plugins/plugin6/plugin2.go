package plugin6

import (
	"context"
)

type Yo interface {
	Yo() string
}

type Plugin2 struct {
}

func (p *Plugin2) Init(super Yo) error {
	println(super.Yo())
	return nil
}

func (p *Plugin2) Serve() chan error {
	errCh := make(chan error, 1)
	return errCh
}

func (p *Plugin2) Stop(context.Context) error {
	return nil
}

func (p *Plugin2) Name() string {
	return "Plugin2"
}
