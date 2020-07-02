package gzip_plugin

import (
	"net/http"

	"github.com/NYTimes/gziphandler"
	cascadeHttp "github.com/spiral/cascade/samples/db_http_logger/modules/http"
	"github.com/spiral/cascade/samples/db_http_logger/modules/logger"
)

type GzipPlugin struct {
	infra  *cascadeHttp.Infrastructure
	logger logger.SuperLogger
}

func (gz *GzipPlugin) Init(i *cascadeHttp.Infrastructure, logger logger.SuperLogger) error {
	gz.infra = i
	gz.logger = logger
	return nil
}

func (gz *GzipPlugin) Serve() chan error {
	errCh := make(chan error)
	return errCh
}

func (gz *GzipPlugin) Stop() error {
	return nil
}

func (gz *GzipPlugin) Configure() error {
	gz.logger.SuperLogToStdOut("added gzip middleware")
	gz.infra.AddMiddleware(gz.middleware)
	return nil
}

func (gz *GzipPlugin) Close() error {
	return nil
}

func (gz *GzipPlugin) middleware(f http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		gziphandler.GzipHandler(f).ServeHTTP(w, r)
	}
}
