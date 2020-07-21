package foo3

import (
	"github.com/spiral/cascade/tests/foo2"
	"github.com/spiral/cascade/tests/foo4"
)

type S3ServeError struct {
}

func (s3 *S3ServeError) Depends() []interface{} {
	return []interface{}{
		s3.SomeOtherDep,
	}
}

func (s3 *S3ServeError) SomeOtherDep(svc *foo4.S4ServeError, svc2 foo2.S2) error {
	return nil
}

// Depends on S3
func (s3 *S3ServeError) Init(svc foo2.S2) error {
	return nil
}

func (s3 *S3ServeError) Serve() chan error {
	errCh := make(chan error, 1)
	return errCh
}

func (s3 *S3ServeError) Stop() error {
	return nil
}
