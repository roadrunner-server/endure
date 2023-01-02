package plugin2

import (
	"github.com/roadrunner-server/endure/v2/tests/disabled_vertices/plugin1"
)

type Plugin2 struct {
}

func (p *Plugin2) Init(p1 plugin1.Plugin1) error {
	return nil
}
