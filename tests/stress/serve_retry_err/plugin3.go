package serve_retry_err

import (
	"errors"
	"math/rand"
)

type S3 struct {
}

func (s3 *S3) Depends() []interface{} {
	return []interface{}{
		s3.SomeOtherDep,
	}
}

func (s3 *S3) SomeOtherDep(svc *S4, svc2 S2) error {
	return nil
}

// Depends on S3
func (s3 *S3) Init(svc S2) error {
	return nil
}

func (s3 *S3) Serve() chan error {
	errCh := make(chan error, 1)
	return errCh
}

func (s3 *S3) Stop() error {
	return nil
}

type S3Init struct {
}

func (s3 *S3Init) Depends() []interface{} {
	return []interface{}{
		s3.SomeOtherDep,
	}
}

func (s3 *S3Init) SomeOtherDep(svc *S4, svc2 S2) error {
	return nil
}

// Depends on S3
func (s3 *S3Init) Init(svc S2) error {
	s := rand.Intn(10)
	// every 5th
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
