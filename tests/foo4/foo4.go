package foo4

import "github.com/spiral/cascade/tests/foo5"

type S4 struct {
}

type DB struct {
	Name string
}

// No deps
func (s *S4) Init(wr foo5.S5) error {
	wr.WRead()
	println("hello from S4 --> Init")
	return nil
}

// But provide some
func (s *S4) Provides() []interface{} {
	return []interface{}{
		s.CreateAnotherDb,
	}
}

// this is the same type but different packages
func (s *S4) CreateAnotherDb() (*DB, error) {
	println("hello from S4 --> CreateAnotherDb")
	return &DB{
		Name: "S4 greeting you, padavan",
	}, nil
}

func (s *S4) Configure() chan error {
	errCh := make(chan error, 1)
	println("S4: configuring")
	go func() {
		errCh <- nil
	}()
	return errCh
}

func (s *S4) Serve() chan error {
	errCh := make(chan error, 1)
	println("S4: serving")
	go func() {
		errCh <- nil
	}()
	return errCh
}

func (s *S4) Close() error {
	println("S4: closing")
	return nil
}

func (s *S4) Stop() error {
	println("S4: stopping")
	return nil
}
