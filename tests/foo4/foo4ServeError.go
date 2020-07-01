package foo4

import (
	"errors"
	"time"

	"github.com/spiral/cascade/tests/foo5"
)

type S4ServeError struct {
}


// No deps
func (s *S4ServeError) Init(wr foo5.S5) error {
	wr.WRead()
	return nil
}

// But provide some
func (s *S4ServeError) Provides() []interface{} {
	return []interface{}{
		s.CreateAnotherDb,
	}
}

// this is the same type but different packages
func (s *S4ServeError) CreateAnotherDb() (*DB, error) {
	return &DB{
		Name: "",
	}, nil
}

func (s *S4ServeError) Configure() error {
	return nil
}

func (s *S4ServeError) Serve() chan error {
	errCh := make(chan error, 1)
	go func() {
		time.Sleep(time.Millisecond * 500)
		errCh <- errors.New("s4 test error")
	}()
	return errCh
}

func (s *S4ServeError) Close() error {
	return nil
}

func (s *S4ServeError) Stop() error {
	return nil
}
