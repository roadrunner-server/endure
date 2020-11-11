package ServeErr

import "github.com/spiral/errors"

type FOO4DB struct {
	Name string
}

type S4ServeError struct {
}

// No deps
func (s *S4ServeError) Init(s5 *S5) error {
	return nil
}

// But provide some
func (s *S4ServeError) Provides() []interface{} {
	return []interface{}{
		s.CreateAnotherDB,
	}
}

// this is the same type but different packages
func (s *S4ServeError) CreateAnotherDB() (*FOO4DB, error) {
	return &FOO4DB{
		Name: "FOO4DB",
	}, nil
}

func (s *S4ServeError) Serve() chan error {
	errCh := make(chan error, 1)
	errCh <- errors.E(errors.Op("S4Serve"), errors.Serve, errors.Str("s4 test error"))
	return errCh
}

func (s *S4ServeError) Stop() error {
	return nil
}
