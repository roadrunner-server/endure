package primitive

import (
	"context"
)

type Plugin8 struct {
}

// Collects on S2 and DB (S3 in the current case)
func (f *Plugin8) Init(a int) error {
	return nil
}

func (f *Plugin8) Serve() chan error {
	errCh := make(chan error, 1)
	return errCh
}

func (f *Plugin8) Stop(context.Context) error {
	return nil
}
