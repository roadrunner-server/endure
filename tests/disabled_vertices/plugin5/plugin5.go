package plugin5

type Plugin5 struct {
}

func (p5 *Plugin5) Init() error {
	return nil
}

func (p5 *Plugin5) FightWithEvil() string {
	return "plugin5 is ready to fight"
}
