package plugin2

import (
	"context"

	"github.com/roadrunner-server/endure/v2/dep"
)

type DB struct {
}

func (d *DB) P2DB() {}

type S2 struct {
}

type P2DB interface {
	P2DB()
}

type P4DB interface {
	P4DB()
}

func (s2 *S2) Init(P4DB) error {
	return nil
}

func (s2 *S2) Provides() []*dep.Out {
	return []*dep.Out{
		dep.OutType((*P2DB)(nil), s2.CreateDB),
	}
}

func (s2 *S2) CreateDB() *DB {
	return &DB{}
}

func (s2 *S2) Serve() chan error {
	errCh := make(chan error, 1)
	return errCh
}

func (s2 *S2) Stop(context.Context) error {
	return nil
}

func (s2 *S2) S2SomeMethod() {}
