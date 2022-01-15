package plugin2

import (
	"github.com/roadrunner-server/endure/tests/issues/issue84/one_alive/plugin3"
	"github.com/roadrunner-server/errors"
)

type Plugin2 struct{}

func (p *Plugin2) Init(_ plugin3.Fooer, _ *plugin3.Plugin3) error {
	return errors.E(errors.Disabled)
}

func (p *Plugin2) Serve() chan error {
	errCh := make(chan error, 1)
	return errCh
}

func (p *Plugin2) Stop() error {
	return nil
}
