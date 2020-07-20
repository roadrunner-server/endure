package named_registers

type Foo10 struct {

}

func (f *Foo10) Init(db DB) error {
	println(db.Name)
	return nil
}

func (f *Foo10) Serve() chan error {
	errCh := make(chan error)
	return errCh
}

func (f *Foo10) Stop() error {
	return nil
}


func (f *Foo10) Name() string{
	return "FOOOOO10"
}
