package plugin6

import "github.com/spiral/endure"

type SuperInterface interface {
	Yo() string
}

type SomeOtherStruct struct {
	name string
}

func (s *SomeOtherStruct) Yo() string {
	return "Yo!"
}

func NewSomeOtherStruct() SuperInterface {
	return &SomeOtherStruct{}
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

func (p *Plugin) Stop() error {
	return nil
}

func (p *Plugin) ProvideWithName(named endure.Named) (SuperInterface, error) {
	println("this is the case, when we need the name")
	println(named.Name())
	return NewSomeOtherStruct(), nil
}

func (p *Plugin) ProvideWithInterfaceAndStruct(named endure.Named, p3 *Plugin3) (SuperInterface, error) {
	println("this is the case, when we need the name and struct")
	println(p3.Boo())
	return NewSomeOtherStruct(), nil
}

func (p *Plugin) ProvideWithOutName() (SuperInterface, error) {
	println("this is the case, when we don't need the name")
	return NewSomeOtherStruct(), nil
}

// Provides declares factory methods.
func (p *Plugin) Provides() []interface{} {
	return []interface{}{
		p.ProvideWithOutName,
		p.ProvideWithName,
		p.ProvideWithInterfaceAndStruct,
	}
}
