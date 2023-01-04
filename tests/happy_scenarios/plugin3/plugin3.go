package plugin3

import (
	"context"

	"github.com/roadrunner-server/endure/v2/dep"
)

type S3 struct {
}

func (s3 *S3) Collects() []*dep.In {
	return []*dep.In{
		dep.Fits(func(p any) {

		}, (*S4Dep)(nil)),
		dep.Fits(func(p any) {

		}, (*S2Dep)(nil)),
	}
}

type S4Dep interface {
	S4SomeMethod()
}

type S2Dep interface {
	S2SomeMethod()
}

func (s3 *S3) Init(S2Dep) error {
	return nil
}

func (s3 *S3) Serve() chan error {
	errCh := make(chan error, 1)
	return errCh
}

func (s3 *S3) Stop(context.Context) error {
	return nil
}
