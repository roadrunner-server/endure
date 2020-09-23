package happyScenarioGraph

import (
	"github.com/spiral/endure/tests/foo5"
)

type S4 struct {
}

type DB struct {
	Name string
}

// No deps
func (s *S4) Init(foo5 foo5.S5) error {
	return nil
}

// But provide some
func (s *S4) Provides() []interface{} {
	return []interface{}{
		s.CreateAnotherDB,
	}
}

// this is the same type but different packages
func (s *S4) CreateAnotherDB() (*DB, error) {
	return &DB{
		Name: "",
	}, nil
}

func (s *S4) Depends() []interface{} {
	return []interface{}{
		s.AddService,
	}
}

func (s *S4) AddService(svc foo5.S5) error {
	return nil
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
