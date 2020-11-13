package plugin3

import "github.com/spiral/endure"

// TODO algo to correctly fill the deps
type Plugin3 struct {
}

type Plugin3Dep struct {
	Name string
}

type Plugin3OtherType struct {
	Name string
}

func (p *Plugin3) Init() error {
	return nil
}

func (p *Plugin3) Serve() chan error {
	errCh := make(chan error, 1)
	return errCh
}

func (p *Plugin3) Stop() error {
	return nil
}

func (p *Plugin3) Provides() []interface{} {
	return []interface{}{
		p.AddDB,
		p.AddDBWithErr,
		p.OtherType,
		p.JustOtherType,
	}
}

func (p *Plugin3) AddDB() *Plugin3Dep {
	return &Plugin3Dep{Name: "Hey Plugin!"}
}

// error will be filtered out
func (p *Plugin3) AddDBWithErr() (*Plugin3Dep, error) {
	return &Plugin3Dep{Name: "Hey Plugin!"}, nil
}

func (p *Plugin3) OtherType(named endure.Named) (*Plugin3Dep, *Plugin3OtherType, error) {
	return &Plugin3Dep{Name: "Hey, I'm with other type"}, &Plugin3OtherType{Name: "Hey, I'm other type"}, nil
}

func (p *Plugin3) JustOtherType() *Plugin3OtherType {
	return &Plugin3OtherType{Name: "Hey, I'm alone here"}
}
