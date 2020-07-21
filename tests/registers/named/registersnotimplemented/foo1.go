package registersnotimplemented

import "github.com/spiral/cascade"

type Foo1 struct {
}

type DB struct {
	Name string
}

func (f *Foo1) Init() error {
	return nil
}

func (f *Foo1) Serve() chan error {
	errCh := make(chan error)
	return errCh
}

func (f *Foo1) Stop() error {
	return nil
}

// But provide some
func (f *Foo1) Provides() []interface{} {
	return []interface{}{
		f.ProvideDB,
	}
}

// this is the same type but different packages
// foo10 invokes foo11
// foo11 should get the foo10 name or provide vertex id
func (f *Foo1) ProvideDB(name cascade.Named) (*DB, error) {
	return &DB{
		Name: name.Name(),
	}, nil
}
