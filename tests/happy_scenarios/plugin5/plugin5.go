package plugin5

import (
	"context"
)

type S5 struct{}

type FOO5DB struct {
	Name string
}

type IFOO5DB interface {
	FOO5DB()
}

func (s *S5) FOO5DB() {}

// No deps
func (s *S5) Init() error {
	return nil
}

func (s *S5) Serve() chan error {
	errCh := make(chan error, 1)
	return errCh
}

func (s *S5) Stop(context.Context) error {
	return nil
}
