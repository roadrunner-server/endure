package plugin9

type P6Dep interface {
	SomeProvidesP6()
}

type Plugin9 struct {
}

func (p9 *Plugin9) Init(P6Dep) error {
	return nil
}
