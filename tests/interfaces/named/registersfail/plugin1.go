package registersfail

type Plugin1 struct {
}

func (f *Plugin1) Init(db *DB) error {
	println(db.Name)
	return nil
}

func (f *Plugin1) Serve() chan error {
	errCh := make(chan error)
	return errCh
}

func (f *Plugin1) Stop() error {
	return nil
}
