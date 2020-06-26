package foo1

import (
	"errors"

	"github.com/spiral/cascade/tests/foo2"
	"github.com/spiral/cascade/tests/foo4"
)

type S1Err struct {
}

type DB struct {
	Name string
}

// No deps
func (s *S1Err) Init(s2 *foo2.S2, db *foo4.DB) error {
	println("hello from S1_err --> Init")
	return nil
}

func (s *S1Err) AddService(svc *foo4.S4) error {
	println("hello from S1_err --> AddService")
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
	println("S1_err: stopping")
	return nil
}
