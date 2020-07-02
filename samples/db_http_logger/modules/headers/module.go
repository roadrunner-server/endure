package headers

import (
	"net/http"

	cascadeHttp "github.com/spiral/cascade/samples/db_http_logger/modules/http"
	"github.com/spiral/cascade/samples/db_http_logger/modules/logger"
)

type Headers struct {
	infra  *cascadeHttp.Infrastructure
	logger logger.SuperLogger
}

func (h *Headers) Init(i *cascadeHttp.Infrastructure, logger logger.SuperLogger) error {
	h.infra = i
	h.logger = logger
	return nil
}

func (h *Headers) Serve() chan error {
	errCh := make(chan error)
	return errCh
}

func (h *Headers) Stop() error {
	return nil
}

func (h *Headers) Configure() error {
	h.logger.SuperLogToStdOut("added headers middleware")
	h.infra.AddMiddleware(h.middleware)
	return nil
}

func (h *Headers) Close() error {
	return nil
}

func (h *Headers) middleware(f http.HandlerFunc) http.HandlerFunc {
	// Define the http.HandlerFunc
	return func(w http.ResponseWriter, r *http.Request) {
		// just dumb function
		f(w, r)
	}
}