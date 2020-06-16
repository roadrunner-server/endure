package foo1

import "errors"

type S1Err struct {
}

type DB struct {
	Name string
}

// No deps
func (s *S1Err) Init() error {
	println("hello from S1 --> Init")
	return errors.New("test error")
}

func (s *S1Err) Serve(upstream chan interface{}) error {
	return nil
}

func (s *S1Err) Stop() error {
	println("S1: error occurred, stopping")
	return nil
}
