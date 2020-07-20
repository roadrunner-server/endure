package named_registers

type Foo11 struct {

}

func (f *Foo11) Init() error {
	return nil
}

func (f *Foo11) Serve() chan error {
	errCh := make(chan error)
	return errCh
}

func (f *Foo11) Stop() error {
	return nil
}

