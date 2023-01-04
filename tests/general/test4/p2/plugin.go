package p2

import (
	"context"

	"github.com/roadrunner-server/endure/v2/dep"
)

type Plugin struct {
	count int
}

type Fooer interface {
	FooBar() string
	Init(val string) error
}

func (p *Plugin) Init() error {
	p.count = 0
	return nil
}

func (p *Plugin) Serve() chan error {
	return make(chan error, 1)
}

func (p *Plugin) Stop(context.Context) error {
	if p.count != 2 {
		panic("cound != 2")
	}
	return nil
}

func (p *Plugin) Name() string {
	return "p2"
}

func (p *Plugin) Collects() []*dep.In {
	return []*dep.In{
		dep.Fits(func(any) {
			p.count++
		}, (*Fooer)(nil)),
	}
}
