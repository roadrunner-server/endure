package plugin7

type P6Dep interface {
	SomeProvidesP6()
}

type Plugin7 struct{}

func (p7 *Plugin7) Init(P6Dep) error {
	return nil
}
