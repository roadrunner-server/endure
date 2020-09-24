package ServeErr

import (
	"errors"
	"time"
)

type S1ServeErr struct {
}

// No deps
func (s *S1ServeErr) Init(s2 *S2, db *DB) error {
	return nil
}

func (s *S1ServeErr) AddService(svc *S4ServeError) error {
	return nil
}

func (s *S1ServeErr) Serve() chan error {
	errCh := make(chan error, 1)
	go func() {
		time.Sleep(time.Millisecond * 1000)
		errCh <- errors.New("test serve error")
	}()
	return errCh
}

func (s *S1ServeErr) Stop() error {
	return nil
}
