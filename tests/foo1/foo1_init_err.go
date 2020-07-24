package foo1

import (
	"errors"

	"github.com/spiral/endure/tests/foo2"
)

type S1Err struct {
}

// No deps
func (s *S1Err) Init(s2 *foo2.S2Err) error {
	return errors.New("s1 test init error")
}

func (s *S1Err) Serve() chan error {
	errCh := make(chan error, 1)
	return errCh
}

func (s *S1Err) Stop() error {
	return nil
}
