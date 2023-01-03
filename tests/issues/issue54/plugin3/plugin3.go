package plugin3

import (
	"context"

	"github.com/roadrunner-server/endure/v2/dep"
)

type Plugin3 struct {
}

type Plugin3Dep struct {
	N string
}

func (pp *Plugin3Dep) Plugin3DepM() {}
func (pp *Plugin3Dep) Name() string {
	return pp.N
}

type IPlugin3Dep interface {
	Plugin3DepM()
}

type Plugin3OtherType struct {
	N string
}

func (pp *Plugin3OtherType) Name() string {
	return pp.N
}

func (pp *Plugin3OtherType) Plugin3OtherDepM() {}

type IPlugin3OtherDepM interface {
	Plugin3OtherDepM()
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
		dep.OutType((*IPlugin3Dep)(nil), p.AddDB),
		dep.OutType((*IPlugin3OtherDepM)(nil), p.OtherType),
	}
}

func (p *Plugin3) AddDB() *Plugin3Dep {
	return &Plugin3Dep{N: "Hey Plugin!"}
}

func (p *Plugin3) OtherType() *Plugin3OtherType {
	return &Plugin3OtherType{N: "Hey, I'm other type"}
}
