package plugin2

import (
	"time"

	"github.com/roadrunner-server/endure/v2/tests/issues/issue55/plugin3"
	"github.com/roadrunner-server/errors"
)

type Plugin2 struct {
}

func (p *Plugin2) Init(p3 *plugin3.Plugin3) error {
	return nil
}

func (p *Plugin2) Serve() chan error {
	errCh := make(chan error, 1)
	time.Sleep(time.Second * 5)
	errCh <- errors.Str("test error from Plugin2")
	return errCh
}

func (p *Plugin2) Stop() error {
	return nil
}
