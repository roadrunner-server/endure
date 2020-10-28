package plugin4

import "net/http"

type Plugin4 struct {
}

func (f8 *Plugin4) Init() error {
	return nil
}

func (f8 *Plugin4) Serve() chan error {
	errCh := make(chan error)
	return errCh
}

func (f8 *Plugin4) Stop() error {
	return nil
}

func (f8 *Plugin4) AddMiddleware(handler http.Handler) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		handler.ServeHTTP(writer, request)
	}
}
