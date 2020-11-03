package ServeRetryErr

import (
	"time"

	"github.com/spiral/errors"
)

type S1ServeErr struct {
}

// No deps
func (s *S1ServeErr) Init(s2 *S2) error {
	return nil
}

func (s *S1ServeErr) Serve() chan error {
	var op = errors.Op("S1 Serve")
	errCh := make(chan error, 1)
	go func() {
		time.Sleep(time.Second)
		err := errors.E(op, errors.Serve, "test serve error")
		errCh <- err
	}()
	return errCh
}

func (s *S1ServeErr) Stop() error {
	return nil
}
