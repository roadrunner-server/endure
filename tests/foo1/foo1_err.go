package foo1

import (
	"errors"
)

type S1Err struct {
}

type DB struct {
	Name string
}

// No deps
func (s *S1Err) Init() error {
	println("hello from S1_err --> Init")
	return nil
}

func (s *S1Err) Serve() chan error {
	errCh := make(chan error, 1)
	println("S1_err: serving")
	go func() {
		errCh <- errors.New("test error")
	}()
	return errCh
}

func (s *S1Err) Stop() error {
	println("S1_err: error occurred, stopping")
	return nil
}
