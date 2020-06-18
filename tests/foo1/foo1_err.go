package foo1

import (
	"errors"
	"time"
)

type S1Err struct {
}

type DB struct {
	Name string
}

// No deps
func (s *S1Err) Init() error {
	println("hello from S1 --> Init")
	return nil
}

func (s *S1Err) Serve() chan error {
	errCh := make(chan error, 1)
	go func() {
		time.Sleep(time.Second * 4)
		errCh <- errors.New("test error")
	}()
	return errCh
}

func (s *S1Err) Stop() error {
	println("S1: error occurred, stopping")
	return nil
}
