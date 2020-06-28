package foo1

import (
	"time"

	"github.com/spiral/cascade/tests/foo2"
	"github.com/spiral/cascade/tests/foo4"
)

type S1 struct {
}

func (s1 *S1) Depends() []interface{} {
	return []interface{}{
		s1.AddService,
	}
}

func (s1 *S1) AddService(svc *foo4.S4) error {
	return nil
}

// Depends on S2 and DB (S3 in the current case)
func (s1 *S1) Init(s2 *foo2.S2, db *foo4.DB) error {
	return nil
}

func (s1 *S1) Serve() chan error {
	errCh := make(chan error, 1)
	return errCh
}

func (s1 *S1) Stop() error {
	time.Sleep(time.Second)
	return nil
}
