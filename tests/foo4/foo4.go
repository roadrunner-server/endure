package foo4

import (
	"github.com/spiral/cascade/tests/foo5"
)

type S4 struct {
}

type DB struct {
	Name string
}

// No deps
func (s *S4) Init(wr foo5.S5) error {
	wr.WRead()
	return nil
}

// But provide some
func (s *S4) Provides() []interface{} {
	return []interface{}{
		s.CreateAnotherDb,
	}
}

// this is the same type but different packages
func (s *S4) CreateAnotherDb() (*DB, error) {
	return &DB{
		Name: "",
	}, nil
}

func (s *S4) Configure() error {
	return nil
}

func (s *S4) Serve() chan error {
	errCh := make(chan error, 1)
	return errCh
}

func (s *S4) Close() error {
	return nil
}

func (s *S4) Stop() error {
	return nil
}
