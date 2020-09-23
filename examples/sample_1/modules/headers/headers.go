package headers

import (
	"net/http"
)

type Headers struct {
}

func (h *Headers) Init() error {
	return nil
}

func (h *Headers) Serve() chan error {
	errCh := make(chan error)
	return errCh
}

func (h *Headers) Stop() error {
	return nil
}

func (h *Headers) Middleware(f http.Handler) http.HandlerFunc {
	// Define the http.HandlerFunc
	return func(w http.ResponseWriter, r *http.Request) {
		f.ServeHTTP(w, r)
	}
}

func (h *Headers) Name() string {
	return "super Headers middleware"
}
