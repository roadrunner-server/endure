package plugin6

type Plugin2 struct {
}

func (p *Plugin2) Init(super SuperInterface) error {
	println(super.Yo())
	return nil
}

func (p *Plugin2) Serve() chan error {
	errCh := make(chan error, 1)
	return errCh
}

func (p *Plugin2) Stop() error {
	return nil
}

func (p *Plugin2) Name() string {
	return "Plugin2"
}
