package issue445_test

type Plugin1 struct {
}

func (p *Plugin1) Init() error {
	return nil
}

func (p *Plugin1) Serve() chan error {
	errCh := make(chan error, 1)
	return errCh
}

func (p *Plugin1) Stop() error {
	return nil
}
