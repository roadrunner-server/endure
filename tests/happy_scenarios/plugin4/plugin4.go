package plugin4

import (
	"github.com/spiral/endure/tests/happy_scenarios/plugin5"
	"github.com/spiral/endure/tests/happy_scenarios/plugin6"
)

type S4 struct {
	fooW plugin6.FooWriter
}

type DB struct {
	Name string
}

// No deps
func (s *S4) Init(foo5 plugin5.S5, fooWriter plugin6.FooWriter) error {
	s.fooW = fooWriter
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
		Name: "foo4DB",
	}, nil
}

func (s *S4) Collects() []interface{} {
	return []interface{}{
		s.AddService,
	}
}

func (s *S4) AddService(svc plugin5.S5) error {
	return nil
}

func (s *S4) Serve() chan error {
	errCh := make(chan error, 1)

	s.fooW.Fooo()

	return errCh
}

func (s *S4) Stop() error {
	return nil
}
