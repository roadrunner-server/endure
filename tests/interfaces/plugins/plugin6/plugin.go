package plugin6

import (
	"context"

	"github.com/roadrunner-server/endure/v2/dep"
)

type SuperInterface interface {
	Yo() string
}

type SomeOtherStruct struct{}

func (s *SomeOtherStruct) Yo() string {
	return "Yo!"
}

func NewSomeOtherStruct() SuperInterface {
	return &SomeOtherStruct{}
}

type Named interface {
	// Name return user friendly name of the plugin
	Name() string
}

type Plugin struct {
}

func (p *Plugin) Init() error {
	return nil
}

func (p *Plugin) Serve() chan error {
	errCh := make(chan error, 1)
	return errCh
}

func (p *Plugin) Stop(context.Context) error {
	return nil
}

// Provides declares factory methods.
func (p *Plugin) Provides() []*dep.Out {
	return []*dep.Out{
		dep.Bind((*SuperInterface)(nil), p.ProvideWithName),
		dep.Bind((*SuperInterface)(nil), p.ProvideWithInterfaceAndStruct),
		dep.Bind((*SuperInterface)(nil), p.ProvideWithOutName),
	}
}

func (p *Plugin) ProvideWithName() SuperInterface {
	println("this is the case, when we need the name")
	println("first")
	return NewSomeOtherStruct()
}

func (p *Plugin) ProvideWithInterfaceAndStruct() SuperInterface {
	println("this is the case, when we need the name and struct")
	println("second")
	return NewSomeOtherStruct()
}

func (p *Plugin) ProvideWithOutName() SuperInterface {
	println("this is the case, when we don't need the name")
	return NewSomeOtherStruct()
}
