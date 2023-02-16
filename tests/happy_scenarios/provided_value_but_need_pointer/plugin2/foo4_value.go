package plugin2

import (
	"context"

	"github.com/roadrunner-server/endure/v2/dep"
)

type Plugin2 struct {
}

type DBV struct {
	Name string
}

func (d *DBV) DBV() {}

type DBVp2 interface {
	DBV()
}

// No deps
func (s *Plugin2) Init() error {
	return nil
}

// But provide some
func (s *Plugin2) Provides() []*dep.Out {
	return []*dep.Out{
		dep.Bind((*DBV)(nil), s.CreateAnotherDB),
	}
}

// this is the same type but different packages
func (s *Plugin2) CreateAnotherDB() *DBV {
	return &DBV{
		Name: "",
	}
}

func (s *Plugin2) Serve() chan error {
	errCh := make(chan error, 1)
	return errCh
}

func (s *Plugin2) Stop(context.Context) error {
	return nil
}
