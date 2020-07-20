package named_registers

type Foo10 struct {

}

func (f *Foo10) Init() error {
	return nil
}

func (f *Foo10) Serve() chan error {
	errCh := make(chan error)
	return errCh
}

func (f *Foo10) Stop() error {
	return nil
}

