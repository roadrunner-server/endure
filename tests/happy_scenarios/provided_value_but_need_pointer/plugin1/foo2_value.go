package plugin1

import "github.com/spiral/endure/tests/happy_scenarios/provided_value_but_need_pointer/plugin2"

type Plugin1 struct {
}

func (s2 *Plugin1) Init(db *plugin2.DBV) error {
	return nil
}

func (s2 *Plugin1) Close() error {
	return nil
}

func (s2 *Plugin1) Configure() error {
	return nil
}

func (s2 *Plugin1) Serve() chan error {
	errCh := make(chan error, 1)
	return errCh
}

func (s2 *Plugin1) Stop() error {
	return nil
}
