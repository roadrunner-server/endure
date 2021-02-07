package plugin1

import (
	"github.com/spiral/endure/tests/issues/issue84/one_alive/plugin3"
)

type Plugin1 struct {
}

func (p *Plugin1) Init(p3 plugin3.Fooer) error {
	println(p3.Foo())
	return nil
}

func (p *Plugin1) Serve() chan error {
	errCh := make(chan error, 1)
	return errCh
}

func (p *Plugin1) Stop() error {
	return nil
}
