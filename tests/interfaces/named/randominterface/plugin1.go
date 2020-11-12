package randominterface

type Plugin1 struct {
}

type SuperInterface interface {
	Super() string
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

func (f *Plugin1) Super() string {
	return "SUPER -> "
}
