package plugin3

import (
	"github.com/spiral/endure/tests/happy_scenarios/plugin2"
	"github.com/spiral/endure/tests/happy_scenarios/plugin4"
)

type S3 struct {
}

func (s3 *S3) Depends() []interface{} {
	return []interface{}{
		s3.SomeOtherDep,
	}
}

func (s3 *S3) SomeOtherDep(svc *plugin4.S4, svc2 plugin2.S2) error {
	return nil
}

// Depends on S3
func (s3 *S3) Init(svc plugin2.S2) error {
	return nil
}

func (s3 *S3) Serve() chan error {
	errCh := make(chan error, 1)
	return errCh
}

func (s3 *S3) Stop() error {
	return nil
}
