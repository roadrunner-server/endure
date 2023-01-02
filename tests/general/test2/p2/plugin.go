package p2

import (
	"github.com/roadrunner-server/endure/v2/dep"
)

type Plugin struct {
}

type Fooer interface {
	FooBar() string
	Init(val string) error
}

type Barer interface {
	bar()
}

func (p *Plugin) Init() error {
	return nil
}

func (p *Plugin) Serve() chan error {
	return make(chan error, 1)
}

func (p *Plugin) Stop() error {
	return nil
}

func (p *Plugin) Name() string {
	return "p2"
}

func callback(p any) {
	pp := p.(Fooer)
	err := pp.Init("booooo")
	if err != nil {
		panic(err)
	}

	println(pp.FooBar())
}

func (p *Plugin) Collects() []*dep.In {
	return []*dep.In{
		dep.Fits(callback, (*Fooer)(nil)),
		//dep.Fits(callback, (*Fooer)(nil)),
	}
}
