package randominterface

type Plugin2 struct {
}

type DB struct {
	Name string
}

func (f *Plugin2) Init() error {
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
	}
}

// this is the same type but different packages
// foo10 invokes foo11
// foo11 should get the foo10 name or provide vertex id
func (f *Plugin2) ProvideDB(super SuperInterface) (*DB, error) {
	return &DB{
		Name: super.Super() + "ME",
	}, nil
}
