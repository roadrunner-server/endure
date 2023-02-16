package ServeErr

import (
	"context"
)

type S5 struct {
}

type FOO5DB struct {
	Name string
}

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

func (s *S5) S5Dep() {}
