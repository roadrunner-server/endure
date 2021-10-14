package plugin5

import (
	"github.com/spiral/errors"
)

type Plugin5 struct {
	val int
}

func (p5 *Plugin5) Init() error {
	p5.val++
	if p5.val > 1 {
		return errors.E("should not be more than 1")
	}
	return nil
}

func (p5 *Plugin5) FightWithEvil() string {
	return "plugin5 is ready to fight"
}
