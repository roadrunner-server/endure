package plugin4

import (
	"context"

	"github.com/roadrunner-server/endure/v2/dep"
)

type S4 struct {
	fooW FooWriter
}

type FooWriter interface {
	Fooo() // just stupid name
}

type DB struct {
	Name string
}

func (d *DB) P4DB() {}

type P4DB interface {
	P4DB()
}

func (s *S4) Init(_ IFOO5DB, fooWriter FooWriter) error {
	s.fooW = fooWriter
	return nil
}

func (s *S4) Provides() []*dep.Out {
	return []*dep.Out{
		dep.OutType((*P4DB)(nil), s.CreateAnotherDB),
	}
}

// this is the same type but different packages
func (s *S4) CreateAnotherDB() *DB {
	return &DB{
		Name: "foo4DB",
	}
}

func (s *S4) Collects() []*dep.In {
	return []*dep.In{
		dep.Fits(func(p any) {

		}, (*IFOO5DB)(nil)),
	}
}

type IFOO5DB interface {
	FOO5DB()
}

func (s *S4) Serve() chan error {
	errCh := make(chan error, 1)

	s.fooW.Fooo()

	return errCh
}

func (s *S4) Stop(context.Context) error {
	return nil
}

func (s *S4) S4SomeMethod() {}
