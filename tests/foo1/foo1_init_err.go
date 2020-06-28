package foo1

import (
	"errors"
	"time"

	"github.com/spiral/cascade/tests/foo2"
	"github.com/spiral/cascade/tests/foo4"
)

type S1Err struct {
}


// No deps
func (s *S1Err) Init(s2 *foo2.S2, db *foo4.DB) error {
	return errors.New("s1 test init error")
}

func (s *S1Err) AddService(svc *foo4.S4) error {
	return nil
}

func (s *S1Err) Serve() chan error {
	errCh := make(chan error, 1)
	return errCh
}

func (s *S1Err) Stop() error {
	time.Sleep(time.Second)
	return nil
}