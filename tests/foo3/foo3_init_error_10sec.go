package foo3

import (
	"errors"
	"math/rand"

	"github.com/spiral/cascade/tests/foo2"
	"github.com/spiral/cascade/tests/foo4"
)

type S3Init struct {
}

func (s3 *S3Init) Depends() []interface{} {
	return []interface{}{
		s3.SomeOtherDep,
	}
}

func (s3 *S3Init) SomeOtherDep(svc *foo4.S4, svc2 foo2.S2) error {
	return nil
}

// Depends on S3
func (s3 *S3Init) Init(svc foo2.S2) error {
	s := rand.Intn(10)
	// just random
	println("---------------------------------------------> " + string(s))
	if s == 5 {
		return errors.New("random error during init from S3")
	}
	return nil
}

func (s3 *S3Init) Serve() chan error {
	errCh := make(chan error, 1)
	return errCh
}

func (s3 *S3Init) Stop() error {
	return nil
}
