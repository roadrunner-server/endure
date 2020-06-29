package foo4

type S4V struct {
}

type DBV struct {
	Name string
}

// No deps
func (s *S4V) Init() error {
	return nil
}

// But provide some
func (s *S4V) Provides() []interface{} {
	return []interface{}{
		s.CreateAnotherDb,
	}
}

// this is the same type but different packages
func (s *S4V) CreateAnotherDb() (DBV, error) {
	return DBV{
		Name: "",
	}, nil
}

func (s *S4V) Configure() error {
	return nil
}

func (s *S4V) Serve() chan error {
	errCh := make(chan error, 1)
	return errCh
}

func (s *S4V) Close() error {
	return nil
}

func (s *S4V) Stop() error {
	return nil
}
