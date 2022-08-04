package collects_get_all_deps

import "github.com/roadrunner-server/errors"

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

func (f *Plugin2) Stop() error {
	return nil
}

func (f *Plugin2) Collects() []any {
	return []any{
		f.GetSuper,
		f.GetSuper2,
	}
}

func (f *Plugin2) GetSuper(s SuperInterface) {
	f.collectsDeps = append(f.collectsDeps, s)
}

func (f *Plugin2) GetSuper2(s Super2Interface) {
	f.collectsDeps = append(f.collectsDeps, s)
}
