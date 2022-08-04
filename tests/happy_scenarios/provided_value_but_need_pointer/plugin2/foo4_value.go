package plugin2

type Plugin2 struct {
}

type DBV struct {
	Name string
}

// No deps
func (s *Plugin2) Init() error {
	return nil
}

// But provide some
func (s *Plugin2) Provides() []any {
	return []any{
		s.CreateAnotherDB,
	}
}

// this is the same type but different packages
func (s *Plugin2) CreateAnotherDB() (DBV, error) {
	return DBV{
		Name: "",
	}, nil
}

func (s *Plugin2) Serve() chan error {
	errCh := make(chan error, 1)
	return errCh
}

func (s *Plugin2) Stop() error {
	return nil
}
