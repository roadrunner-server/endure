package registers

type Plugin1 struct {
}

func (f *Plugin1) Init(db *DB, db2 *DB2) error {
	println(db.Name)
	println(db2.Name)
	return nil
}

func (f *Plugin1) Serve() chan error {
	errCh := make(chan error)
	return errCh
}

func (f *Plugin1) Stop() error {
	return nil
}

func (f *Plugin1) Name() string {
	return "My name is Plugin1, friend!"
}
