package registers

import endure "github.com/spiral/endure/pkg/container"

type Plugin2 struct {
}

type DB struct {
	Name string
}

type DB2 struct {
	Name string
}

func (f *Plugin2) Init() error {
	println("plugin2 Init called")
	return nil
}

func (f *Plugin2) Serve() chan error {
	errCh := make(chan error)
	return errCh
}

func (f *Plugin2) Stop() error {
	return nil
}

// But provide some
func (f *Plugin2) Provides() []interface{} {
	return []interface{}{
		f.ProvideDB,
		f.ProvideDB2,
	}
}

// this is the same type but different packages
// foo10 invokes foo11
// foo11 should get the foo10 name or provide vertex id
func (f *Plugin2) ProvideDB(name endure.Named) (*DB, error) {
	println("ProvideDB called")
	return &DB{
		Name: name.Name(),
	}, nil
}

// this is the same type but different packages
// foo10 invokes foo11
// foo11 should get the foo10 name or provide vertex id
func (f *Plugin2) ProvideDB2(name endure.Named, name2 endure.Named) (*DB2, error) {
	println("ProvideDB2 called")
	return &DB2{
		Name: name.Name() + "; " + name2.Name(),
	}, nil
}
