package plugin1

import (
	"context"

	"github.com/roadrunner-server/endure/v2/tests/happy_scenarios/provided_value_but_need_pointer/plugin2"
)

type Plugin1 struct {
}

func (s2 *Plugin1) Init(db *plugin2.DBV) error {
	return nil
}

func (s2 *Plugin1) Serve() chan error {
	errCh := make(chan error, 1)
	return errCh
}

func (s2 *Plugin1) Stop(context.Context) error {
	return nil
}
