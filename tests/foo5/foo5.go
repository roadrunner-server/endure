package foo5

import "time"

type Reader interface {
	WRead() // just stupid name
}


type S5 struct {
}

func (s *S5) WRead() {
}

type DB struct {
	Name string
}

// No deps
func (s *S5) Init() error {
	return nil
}

func (s *S5) Configure() error {
	time.Sleep(time.Second)
	return nil
}

func (s *S5) Serve() chan error {
	errCh := make(chan error, 1)
	return errCh
}

func (s *S5) Close() error {
	time.Sleep(time.Second)
	return nil
}

func (s *S5) Stop() error {
	time.Sleep(time.Second)
	return nil
}