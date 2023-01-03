package plugin5

import (
	"context"
	"net/http"

	"github.com/roadrunner-server/endure/v2/dep"
)

type Plugin5 struct {
	mdwr []HTTPMiddleware
}

type HTTPMiddleware interface {
	AddMiddleware(h http.Handler) http.HandlerFunc
}

func (f9 *Plugin5) Init() error {
	return nil
}

func (f9 *Plugin5) Serve() chan error {
	// TEST CHECK
	if len(f9.mdwr) != 2 {
		panic("not enough middlewares")
	}

	errCh := make(chan error)
	return errCh
}

func (f9 *Plugin5) Stop(context.Context) error {
	return nil
}

func (f9 *Plugin5) Collects() []*dep.In {
	return []*dep.In{
		dep.Fits(func(p any) {
			f9.mdwr = append(f9.mdwr, p.(HTTPMiddleware))
		}, (*HTTPMiddleware)(nil)),
	}
}
