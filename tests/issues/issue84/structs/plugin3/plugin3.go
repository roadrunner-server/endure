package plugin3

import "github.com/spiral/errors"

type Plugin3 struct{}

func (p *Plugin3) Init() error {
	return errors.E(errors.Disabled)
}

func (p *Plugin3) Serve() chan error {
	errCh := make(chan error, 1)
	return errCh
}

func (p *Plugin3) Stop() error {
	return nil
}
