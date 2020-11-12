package plugin2

import (
	"github.com/spiral/endure/tests/issues/issue54/plugin3"
)

type Plugin2 struct {
}

func (p *Plugin2) Init(p3 *plugin3.Plugin3Dep, po *plugin3.Plugin3OtherType) error {
	println(p3.Name)
	println(po.Name)
	return nil
}

func (p *Plugin2) Serve() chan error {
	errCh := make(chan error, 1)
	return errCh
}

func (p *Plugin2) Stop() error {
	return nil
}
