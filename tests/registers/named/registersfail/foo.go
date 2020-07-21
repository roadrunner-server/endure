package registersfail

type Foo struct {
}

func (f *Foo) Init(db DB) error {
	println(db.Name)
	return nil
}

func (f *Foo) Serve() chan error {
	errCh := make(chan error)
	return errCh
}

func (f *Foo) Stop() error {
	return nil
}
