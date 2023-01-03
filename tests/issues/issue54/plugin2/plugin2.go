package plugin2

import (
	"context"
)

type Plugin2 struct {
}

type IPlugin3Dep interface {
	Plugin3DepM()
	Name() string
}

type IPlugin3OtherDepM interface {
	Plugin3OtherDepM()
	Name() string
}

func (p *Plugin2) Init(p3 IPlugin3Dep, po IPlugin3OtherDepM) error {
	println(p3.Name())
	println(po.Name())
	return nil
}

func (p *Plugin2) Serve() chan error {
	errCh := make(chan error, 1)
	return errCh
}

func (p *Plugin2) Stop(context.Context) error {
	return nil
}

func (p *Plugin2) SomeMethodP2() {}
