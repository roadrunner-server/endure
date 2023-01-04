package registers

import (
	"context"

	"github.com/roadrunner-server/endure/v2/dep"
)

type Plugin2 struct {
}

type DB struct {
	n string
}

func (d *DB) Select() {}
func (d *DB) Name() string {
	return d.n
}

type IDB interface {
	Select()
	Name() string
}

func (f *Plugin2) Init() error {
	println("plugin2 Init called")
	return nil
}

func (f *Plugin2) Serve() chan error {
	errCh := make(chan error)
	return errCh
}

func (f *Plugin2) Stop(context.Context) error {
	return nil
}

func (f *Plugin2) Provides() []*dep.Out {
	return []*dep.Out{
		dep.Bind((*IDB)(nil), f.ProvideDB),
	}
}

func (f *Plugin2) ProvideDB() *DB {
	println("ProvideDB called")
	return &DB{
		n: "DB",
	}
}
