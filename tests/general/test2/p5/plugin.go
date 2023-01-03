package p5

import (
	"context"

	"github.com/roadrunner-server/endure/v2/dep"
	"github.com/roadrunner-server/endure/v2/tests/general/test1/p1/pkg"
)

type Plugin struct {
	nnn string
}

type Fooer interface {
	FooBar() string
	Init(val string) error
}

func (p *Plugin) Init() error {
	p.nnn = "foo"
	return nil
}

func (p *Plugin) Serve() chan error {
	return make(chan error, 1)
}

func (p *Plugin) Stop(context.Context) error {
	return nil
}

func (p *Plugin) Name() string {
	return "p5"
}

func (p *Plugin) Weight() uint {
	return 110
}

func (p *Plugin) Provides() []*dep.Out {
	return []*dep.Out{
		dep.OutType((*Fooer)(nil), p.InitFoo),
	}
}

func (p *Plugin) InitFoo() *pkg.Foo {
	return pkg.InitFoo()
}

func (p *Plugin) Work() {
	println("wooooorking5!!!")
}
