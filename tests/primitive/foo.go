package primitive

type Foo struct {
}

// Depends on S2 and DB (S3 in the current case)
func (f *Foo) Init(a int) error {
	return nil
}

func (f *Foo) Serve() chan error {
	errCh := make(chan error, 1)
	return errCh
}

func (f *Foo) Stop() error {
	return nil
}
