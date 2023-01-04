package ServeErr

import (
	"context"

	"github.com/roadrunner-server/endure/v2/dep"
)

type DB struct{}

func (d *DB) SuperDB() {}

type S2 struct{}

type SuperDB interface {
	SuperDB()
}

func (s2 *S2) Init(SuperSelecter) error {
	return nil
}

func (s2 *S2) Provides() []*dep.Out {
	return []*dep.Out{
		dep.Bind((*SuperDB)(nil), s2.CreateDB),
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
