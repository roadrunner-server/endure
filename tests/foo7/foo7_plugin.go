package foo7

import "net/http"

type Foo7 struct {
}

func (f7 *Foo7) Init() error {
	return nil
}

func (f7 *Foo7) Serve() chan error {
	errCh := make(chan error)
	return errCh
}

func (f7 *Foo7) Stop() error {
	return nil
}

func (f7 *Foo7) AddMiddleware(handler http.Handler) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		handler.ServeHTTP(writer, request)
	}
}
