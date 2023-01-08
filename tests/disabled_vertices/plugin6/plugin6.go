package plugin6

import (
	"github.com/roadrunner-server/errors"
)

type Plugin6 struct {
}

func (p6 *Plugin6) Init() error {
	return errors.E(errors.Op("plugin6 init"), errors.Disabled)
}

func (p6 *Plugin6) SomeProvidesP6() {}
