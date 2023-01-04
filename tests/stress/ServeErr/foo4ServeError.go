package ServeErr

import (
	"context"

	"github.com/roadrunner-server/endure/v2/dep"
	"github.com/roadrunner-server/errors"
)

type FOO4DB struct {
	Name string
}

func (f4 *FOO4DB) SuperSelect() string {
	return "super-select"
}

type SuperSelecter interface {
	SuperSelect() string
}

type S5Deper interface {
	S5Dep()
}

type S4ServeError struct {
}

// No deps
func (s *S4ServeError) Init(S5Deper) error {
	return nil
}

// But provide some
func (s *S4ServeError) Provides() []*dep.Out {
	return []*dep.Out{
		dep.Bind((*SuperSelecter)(nil), s.CreateAnotherDB),
	}
}

// this is the same type but different packages
func (s *S4ServeError) CreateAnotherDB() *FOO4DB {
	return &FOO4DB{
		Name: "FOO4DB",
	}
}

func (s *S4ServeError) Serve() chan error {
	errCh := make(chan error, 1)
	errCh <- errors.E(errors.Op("S4Serve"), errors.Serve, errors.Str("s4 test error"))
	return errCh
}

func (s *S4ServeError) Stop(context.Context) error {
	return nil
}
