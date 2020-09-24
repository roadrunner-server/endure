package plugin5

import (
	"net/http"
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

func (f9 *Plugin5) Stop() error {
	return nil
}

func (f9 *Plugin5) Depends() []interface{} {
	return []interface{}{
		f9.AddMiddleware,
	}
}

func (f9 *Plugin5) AddMiddleware(m HTTPMiddleware) error {
	f9.mdwr = append(f9.mdwr, m)
	return nil
}
