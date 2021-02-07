package plugin2

import (
	"github.com/spiral/endure/tests/issues/issue84/structs/plugin3"
)

type Plugin2 struct{}

func (p *Plugin2) Init(_ *plugin3.Plugin3) error {
	return nil
}

func (p *Plugin2) Serve() chan error {
	errCh := make(chan error, 1)
	return errCh
}

func (p *Plugin2) Stop() error {
	return nil
}
