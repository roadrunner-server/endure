package plugin5

type Plugin5 struct {
	val int
}

func (p5 *Plugin5) Init() error {
	p5.val++
	println(p5.val)
	return nil
}

func (p5 *Plugin5) FightWithEvil() string {
	return "plugin5 is ready to fight"
}
