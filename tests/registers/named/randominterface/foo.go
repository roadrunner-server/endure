package randominterface

type Foo struct {
}

type SuperInterface interface {
	Super() string
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

func (f *Foo) Super() string {
	return "SUPER -> "
}
