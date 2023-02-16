package plugin4

import "github.com/roadrunner-server/errors"

type Plugin4 struct {
}

func (p4 *Plugin4) Init() error {
	return errors.E(errors.Op("plugin 4 init"), errors.Disabled)
}

func (p4 *Plugin4) FightWithEvil() string {
	return "plugin4 is ready to fight"
}
