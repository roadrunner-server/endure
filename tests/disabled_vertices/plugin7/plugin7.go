package plugin7

import "github.com/roadrunner-server/endure/tests/disabled_vertices/plugin6"

type Plugin7 struct {
}

func (p7 *Plugin7) Init(plugin6 plugin6.Plugin6) error {
	return nil
}
