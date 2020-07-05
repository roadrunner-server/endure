package gzip

import (
	"net/http"

	"github.com/NYTimes/gziphandler"
	cascadeHttp "github.com/spiral/cascade/samples/db_http_logger/modules/http"
	"github.com/spiral/cascade/samples/db_http_logger/modules/logger"
)

type Gzip struct {
	infra  *cascadeHttp.Infrastructure
	logger logger.SuperLogger
}

func (gz *Gzip) Init(i *cascadeHttp.Infrastructure, logger logger.SuperLogger) error {
	logger.SuperLogToStdOut("intializing Gzip")
	gz.infra = i
	gz.logger = logger
	return nil
}

func (gz *Gzip) Serve() chan error {
	gz.logger.SuperLogToStdOut("serving Gzip")
	errCh := make(chan error)
	return errCh
}

func (gz *Gzip) Stop() error {
	return nil
}

func (gz *Gzip) Configure() error {
	gz.logger.SuperLogToStdOut("added gzip middleware")
	gz.infra.AddMiddleware(gz.middleware)
	return nil
}

func (gz *Gzip) Close() error {
	return nil
}

func (gz *Gzip) middleware(f http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		gziphandler.GzipHandler(f).ServeHTTP(w, r)
	}
}
