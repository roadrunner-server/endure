package registers

type Foo10 struct {
}

func (f *Foo10) Init(db DB, db2 DB2) error {
	println(db.Name)
	println(db2.Name)
	return nil
}

func (f *Foo10) Serve() chan error {
	errCh := make(chan error)
	return errCh
}

func (f *Foo10) Stop() error {
	return nil
}

func (f *Foo10) Name() string {
	return "My name is Foo10, friend!"
}
