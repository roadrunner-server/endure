package plugin1

import "github.com/spiral/endure/tests/issues/issue55/plugin2"

type Plugin1 struct {
}

func (p *Plugin1) Init(p2 plugin2.Plugin2) error {
	return nil
}

func (p *Plugin1) Serve() chan error {
	errCh := make(chan error, 1)
	return errCh
}

func (p *Plugin1) Stop() error {
	return nil
}
