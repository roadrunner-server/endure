package CyclicDeps

type Plugin3 struct {
}

func (p3 *Plugin3) Init(p1 *Plugin1) error {
	return nil
}

func (p3 *Plugin3) Serve() chan error {
	errCh := make(chan error, 1)
	return errCh
}

func (p3 *Plugin3) Stop() error {
	return nil
}
