package plugin3

import (
	"github.com/spiral/endure/errors"
)

type Super interface {
	FightWithEvil() string
}

type Plugin3 struct {
}

func (p3 *Plugin3) Init(s Super) error {
	str := s.FightWithEvil()
	if str != "plugin5 is ready to fight" {
		return errors.E(errors.Op("plugin3 init"), errors.Init, errors.Str("no one fight with evil"))
	}
	println(str)
	return nil
}
