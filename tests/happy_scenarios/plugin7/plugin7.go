package plugin7

type Plugin7 struct {
}

func (s1 *Plugin7) Init() error {
	return nil
}

func (s1 *Plugin7) Serve() chan error {
	errCh := make(chan error, 1)
	return errCh
}

func (s1 *Plugin7) Stop() error {
	return nil
}
