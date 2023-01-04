package collects_get_all_deps

import (
	"context"

	"github.com/roadrunner-server/endure/v2/dep"
	"github.com/roadrunner-server/errors"
)

type Plugin2 struct {
	collectsDeps []any
}

func (f *Plugin2) Init() error {
	// should be 2 deps
	f.collectsDeps = make([]any, 0, 2)
	return nil
}

func (f *Plugin2) Serve() chan error {
	errCh := make(chan error)
	if len(f.collectsDeps) != 2 {
		errCh <- errors.E("not enough deps collected")
	}
	return errCh
}

func (f *Plugin2) Stop(context.Context) error {
	return nil
}

func (f *Plugin2) Collects() []*dep.In {
	return []*dep.In{
		dep.Fits(func(p any) {
			f.collectsDeps = append(f.collectsDeps, p)
		}, (*SuperInterface)(nil)),
		dep.Fits(func(p any) {
			f.collectsDeps = append(f.collectsDeps, p)
		}, (*Super2Interface)(nil)),
	}
}
