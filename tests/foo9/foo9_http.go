package foo9

import (
	"net/http"
)

type Foo9 struct {
	mdwr []HttpMiddleware
}

type HttpMiddleware interface {
	AddMiddleware(h http.Handler) http.HandlerFunc
}

func (f9 *Foo9) Init() error {
	return nil
}

func (f9 *Foo9) Serve() chan error {
	// TEST CHECK
	if len(f9.mdwr) != 2 {
		panic("not enough middlewares")
	}

	errCh := make(chan error)
	return errCh
}

func (f9 *Foo9) Stop() error {
	return nil
}

func (f9 *Foo9) Depends() []interface{} {
	return []interface{}{
		f9.AddMiddleware,
	}
}

func (f9 *Foo9) AddMiddleware(m HttpMiddleware) error {
	f9.mdwr = append(f9.mdwr, m)
	return nil
}
