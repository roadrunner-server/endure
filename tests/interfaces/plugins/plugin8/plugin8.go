package plugin8

import (
	endure "github.com/roadrunner-server/endure/pkg/container"
	"github.com/roadrunner-server/endure/tests/interfaces/plugins/plugin10"
	"github.com/roadrunner-server/errors"
)

type SomeInterface interface {
	Boom()
}

type Plugin8 struct {
	collectedDeps []interface{}
}

// No deps
func (s *Plugin8) Init() error {
	s.collectedDeps = make([]interface{}, 0, 6)
	return nil
}

func (s *Plugin8) Serve() chan error {
	errCh := make(chan error, 1)
	// plugin7
	// plugin9
	// named + plugin10 (plugin7)
	// named + plugin10 (plugin9)
	// plugin7 + plugin10
	// plugin9 + plugin10
	if len(s.collectedDeps) != 6 {
		errCh <- errors.E("not enough deps collected")
	}
	return errCh
}

func (s *Plugin8) Name() string {
	return "plugin8"
}

func (s *Plugin8) Collects() []interface{} {
	return []interface{}{
		s.SomeCollects,
		s.SomeCollects2,
		s.SomeCollects3,
	}
}

func (s *Plugin8) SomeCollects(named endure.Named, b SomeInterface, p10 *plugin10.Plugin10) error {
	s.collectedDeps = append(s.collectedDeps, b)
	b.Boom()
	return nil
}

func (s *Plugin8) SomeCollects2(named endure.Named, p10 *plugin10.Plugin10) error {
	s.collectedDeps = append(s.collectedDeps, p10)
	println(named.Name())
	p10.Boo()
	return nil
}

func (s *Plugin8) SomeCollects3(named endure.Named, p10 *plugin10.Plugin10, named2 endure.Named, p *plugin10.Plugin10) error {
	s.collectedDeps = append(s.collectedDeps, p)
	println(named.Name())
	println(named2.Name())
	p10.Boo()
	p.Boo()
	return nil
}

func (s *Plugin8) Stop() error {
	return nil
}
