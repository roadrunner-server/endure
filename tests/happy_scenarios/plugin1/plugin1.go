package plugin1

import (
	"context"

	"github.com/roadrunner-server/endure/v2/dep"
)

type S1 struct {
}

func (s1 *S1) Collects() []*dep.In {
	return []*dep.In{
		dep.Fits(func(p any) {

		}, (*P4DB)(nil)),
	}
}

type P4DB interface {
	P4DB()
}

type P2DB interface {
	P2DB()
}

type S2Dep interface {
	S2SomeMethod()
}

func (s1 *S1) Init(S2Dep, P2DB) error {
	return nil
}

func (s1 *S1) Serve() chan error {
	errCh := make(chan error, 1)
	return errCh
}

func (s1 *S1) Stop(context.Context) error {
	return nil
}
