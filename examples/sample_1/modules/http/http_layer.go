package http

import (
	"context"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/rs/cors"
	"github.com/spiral/cascade/examples/db_http_logger/modules/db"
	"github.com/spiral/cascade/examples/db_http_logger/modules/logger"
)

type Http struct {
	client http.Client
	server *http.Server
	mdwr   []Middleware
	db     db.Repository
	logger logger.SuperLogger
}

// Middleware interface
type Middleware interface {
	Middleware(f http.Handler) http.HandlerFunc
}

func (h *Http) Init(db db.Repository, logger logger.SuperLogger) error {
	logger.SuperLogToStdOut("initializing http")
	tr := &http.Transport{
		MaxIdleConns:       10,
		IdleConnTimeout:    30 * time.Second,
		DisableCompression: true,
	}
	client := http.Client{
		Transport: tr,
		Timeout:   60,
	}
	h.client = client
	h.db = db
	h.logger = logger
	return nil
}

func (h *Http) Serve() chan error {
	h.logger.SuperLogToStdOut("serving http")
	errCh := make(chan error, 1)

	f := h.server.Handler

	// chain middleware
	for i := 0; i < len(h.mdwr); i++ {
		f = h.mdwr[i].Middleware(f)
	}

	h.server.Handler = f

	go func() {
		err := h.server.ListenAndServe()
		if err == http.ErrServerClosed {
			return
		} else {
			errCh <- err
		}
	}()
	return errCh
}

func (h *Http) Stop() error {
	err := h.server.Shutdown(context.Background())
	if err != nil {
		return err
	}
	return nil
}

func (h *Http) Configure() error {
	h.logger.SuperLogToStdOut("configuring http")
	r := mux.NewRouter()

	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowCredentials: true,
		AllowedMethods:   []string{"GET", "HEAD", "POST", "PUT", "OPTIONS"},
		AllowedHeaders:   []string{"*"},
	})

	r.Methods("POST").HandlerFunc(h.update).Path("/update")
	r.Methods("POST").HandlerFunc(h.ddelete).Path("/delete")
	r.Methods("GET").HandlerFunc(h.sselect).Path("/select")
	r.Methods("POST").HandlerFunc(h.insert).Path("/insert")

	// just as sample, we put server here
	server := &http.Server{
		Addr:           ":8080",
		Handler:        c.Handler(r),
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	h.server = server

	return nil
}

func (h *Http) Depends() []interface{} {
	return []interface{}{
		h.AddMiddleware,
	}
}

func (h *Http) AddMiddleware(m Middleware) error {
	h.mdwr = append(h.mdwr, m)
	return nil
}

func (h *Http) Close() error {
	return nil
}

///////////////// INFRA HANDLERS //////////////////////////////

func (h *Http) update(writer http.ResponseWriter, request *http.Request) {
	h.db.Update()
	writer.WriteHeader(http.StatusOK)
}

// ddelete just to not collide with delete keyword
func (h *Http) ddelete(writer http.ResponseWriter, request *http.Request) {
	h.db.Delete()
	writer.WriteHeader(http.StatusOK)
}

// sselect just to not collide with select keyword
func (h *Http) sselect(writer http.ResponseWriter, request *http.Request) {
	h.db.Select()
	writer.WriteHeader(http.StatusOK)

	for i := 0; i < 10000; i++ {
		writer.Write([]byte("TEST_GZIP_HEADERS"))
	}

}
func (h *Http) insert(writer http.ResponseWriter, request *http.Request) {
	h.db.Insert()
	writer.WriteHeader(http.StatusOK)
}

func (h *Http) Name() string {
	return "super http service"
}