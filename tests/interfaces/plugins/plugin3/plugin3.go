package plugin3

import "net/http"

type Plugin3 struct {
}

func (f7 *Plugin3) Init() error {
	return nil
}

func (f7 *Plugin3) Serve() chan error {
	errCh := make(chan error)
	return errCh
}

func (f7 *Plugin3) Stop() error {
	return nil
}

func (f7 *Plugin3) AddMiddleware(handler http.Handler) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		handler.ServeHTTP(writer, request)
	}
}
