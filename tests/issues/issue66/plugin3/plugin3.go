package plugin3

type Plugin3 struct {
}

type Plugin3DB struct {
	Name string
}

func (p *Plugin3) Init() error {
	return nil
}

func (p *Plugin3) Serve() chan error {
	errCh := make(chan error, 1)
	return errCh
}

func (p *Plugin3) Stop() error {
	return nil
}

func (p *Plugin3) Provides() []interface{} {
	return []interface{}{
		p.ProvidePlugin3DB,
	}
}

func (p *Plugin3) ProvidePlugin3DB() *Plugin3DB {
	return &Plugin3DB{
		Name: "plugin3DB",
	}
}
