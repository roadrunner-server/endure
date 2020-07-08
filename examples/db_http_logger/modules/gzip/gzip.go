package gzip

import (
	"net/http"

	"github.com/NYTimes/gziphandler"
)

type Gzip struct {

}

func (gz *Gzip) Init() error {
	return nil
}

func (gz *Gzip) Serve() chan error {
	errCh := make(chan error)
	return errCh
}

func (gz *Gzip) Stop() error {
	return nil
}


func (gz *Gzip) Middleware(f http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		gziphandler.GzipHandler(f).ServeHTTP(w, r)
	}
}
