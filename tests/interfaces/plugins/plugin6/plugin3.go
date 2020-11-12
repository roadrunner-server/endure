package plugin6

type Plugin3 struct {
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

func (p *Plugin3) Boo() string {
	return "Boo from plugin3"
}
