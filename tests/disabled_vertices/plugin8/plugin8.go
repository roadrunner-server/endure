package plugin8

type P6Dep interface {
	SomeProvidesP6()
}

type Plugin8 struct{}

func (p9 *Plugin8) Init(P6Dep) error {
	return nil
}
