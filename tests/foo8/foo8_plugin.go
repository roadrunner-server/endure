package foo8

import "net/http"

type Foo8 struct {

}

func (f8 *Foo8) Init() error {
	return nil
}

func (f8 *Foo8) Serve() chan error {
	errCh := make(chan error)
	return errCh
}

func (f8 *Foo8) Stop() error {
	return nil
}

func (f8 *Foo8) AddMiddleware(handler http.Handler) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		handler.ServeHTTP(writer, request)
	}
}