package foo1

import (
	"errors"
	"time"

	"github.com/spiral/cascade/tests/foo2"
	"github.com/spiral/cascade/tests/foo4"
)

type S1ServeErr struct {
}


// No deps
func (s *S1ServeErr) Init(s2 *foo2.S2, db *foo4.DB) error {
	return nil
}

func (s *S1ServeErr) AddService(svc *foo4.S4) error {
	return nil
}

func (s *S1ServeErr) Serve() chan error {
	errCh := make(chan error, 1)
	go func() {
		time.Sleep(time.Second * 1)
		errCh <- errors.New("s1 test error")
	}()
	return errCh
}

func (s *S1ServeErr) Stop() error {
	time.Sleep(time.Second)
	return nil
}
