package ServeErr

import (
	"context"

	"github.com/roadrunner-server/endure/v2/dep"
)

type S3ServeError struct {
}

func (s3 *S3ServeError) Collects() []*dep.In {
	return []*dep.In{
		dep.Fits(func(p any) {

		}, (*SuperSelecter)(nil)),
	}
}

// Collects on S3
func (s3 *S3ServeError) Init(SuperDB) error {
	return nil
}

func (s3 *S3ServeError) Serve() chan error {
	errCh := make(chan error, 1)
	return errCh
}

func (s3 *S3ServeError) Stop(context.Context) error {
	return nil
}
