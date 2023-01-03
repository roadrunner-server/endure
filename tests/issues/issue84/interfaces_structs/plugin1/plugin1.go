package plugin1

import (
	"context"

	"github.com/roadrunner-server/endure/v2/tests/issues/issue84/interfaces_structs/plugin2"
)

type Plugin1 struct {
}

func (p *Plugin1) Init(_ *plugin2.Plugin2) error {
	return nil
}

func (p *Plugin1) Serve() chan error {
	errCh := make(chan error, 1)
	return errCh
}

func (p *Plugin1) Stop(context.Context) error {
	return nil
}
