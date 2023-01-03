package plugin3

import (
	"context"

	"github.com/roadrunner-server/endure/v2/dep"
)

type Plugin3 struct {
}

type Plugin3DB struct {
	N string
}

func (p *Plugin3DB) SomeDB3DepMethod() {}

func (p *Plugin3DB) Name() string {
	return p.N
}

type IDB3 interface {
	SomeDB3DepMethod()
	Name() string
}

func (p *Plugin3) Init() error {
	return nil
}

func (p *Plugin3) Serve() chan error {
	errCh := make(chan error, 1)
	return errCh
}

func (p *Plugin3) Stop(context.Context) error {
	return nil
}

func (p *Plugin3) Provides() []*dep.Out {
	return []*dep.Out{
		dep.OutType((*IDB3)(nil), p.ProvidePlugin3DB),
	}
}

func (p *Plugin3) ProvidePlugin3DB() *Plugin3DB {
	return &Plugin3DB{
		N: "plugin3DB",
	}
}
