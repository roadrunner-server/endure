package plugin2

import (
	"github.com/spiral/endure/tests/disabled_vertices/plugin1"
)

type Plugin2 struct {
}

func (p *Plugin2) Init(p2 plugin1.Plugin1) error {
	return nil
}