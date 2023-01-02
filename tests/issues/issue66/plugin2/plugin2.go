package plugin2

import (
	"fmt"

	"github.com/roadrunner-server/endure/v2/tests/issues/issue66/plugin3"
)

type Plugin2 struct {
}

func (p *Plugin2) Init(p3 *plugin3.Plugin3DB) error {
	fmt.Println(p3.Name)
	return nil
}

func (p *Plugin2) Serve() chan error {
	errCh := make(chan error, 1)
	return errCh
}

func (p *Plugin2) Stop() error {
	return nil
}
