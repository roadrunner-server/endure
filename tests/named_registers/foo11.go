package named_registers

import "github.com/spiral/cascade"

type Foo11 struct {

}

type DB struct {
	Name string
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

// But provide some
func (f *Foo11) Provides() []interface{} {
	return []interface{}{
		f.ProvideDB,
	}
}

// this is the same type but different packages
func (f *Foo11) ProvideDB(name cascade.Named) (*DB, error) {
	return &DB{
		Name: name.Name(),
	}, nil
}