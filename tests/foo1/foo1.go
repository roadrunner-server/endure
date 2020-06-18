package foo1

import (
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
	println("hello from S1 --> AddService")
	return nil
}

// Depends on S2 and DB (S3 in the current case)
func (s1 *S1) Init(s2 *foo2.S2, db *foo4.DB) error {
	println("hello from S1 --> Init")
	println("S4 in S1: " + db.Name)
	return nil
}

func (s1 *S1) Serve() chan error {
	errCh := make(chan error, 1)
	println("S1: serving")
	go func() {
		errCh <- nil
	}()
	return errCh
}

func (s1 *S1) Stop() error {
	println("S1: stopping")
	return nil
}
