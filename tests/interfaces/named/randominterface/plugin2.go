package randominterface

import (
	"context"

	"github.com/roadrunner-server/endure/v2/dep"
)

type Plugin2 struct {
}

type DB struct {
	Name string
}

func (f *Plugin2) Init() error {
	return nil
}

func (f *Plugin2) Serve() chan error {
	errCh := make(chan error)
	return errCh
}

func (f *Plugin2) Stop(context.Context) error {
	return nil
}

// But provide some
func (f *Plugin2) Provides() []*dep.Out {
	return []*dep.Out{
		dep.OutType((*SuperInterface)(nil), f.ProvideDB),
	}
}

func (f *Plugin2) ProvideDB() *DB {
	return &DB{
		Name: "DB",
	}
}
