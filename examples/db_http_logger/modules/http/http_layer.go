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

type Infrastructure struct {
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

func (infra *Infrastructure) Init(db db.Repository, logger logger.SuperLogger) error {
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
	infra.client = client
	infra.db = db
	infra.logger = logger
	return nil
}

func (infra *Infrastructure) Serve() chan error {
	infra.logger.SuperLogToStdOut("serving http")
	errCh := make(chan error, 1)

	f := infra.server.Handler

	// chain middleware
	for i := 0; i < len(infra.mdwr); i++ {
		f = infra.mdwr[i].Middleware(f)
	}

	infra.server.Handler = f

	go func() {
		err := infra.server.ListenAndServe()
		if err == http.ErrServerClosed {
			return
		} else {
			errCh <- err
		}
	}()
	return errCh
}

func (infra *Infrastructure) Stop() error {
	err := infra.server.Shutdown(context.Background())
	if err != nil {
		return err
	}
	return nil
}

func (infra *Infrastructure) Configure() error {
	infra.logger.SuperLogToStdOut("configuring http")
	r := mux.NewRouter()

	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowCredentials: true,
		AllowedMethods:   []string{"GET", "HEAD", "POST", "PUT", "OPTIONS"},
		AllowedHeaders:   []string{"*"},
	})

	r.Methods("POST").HandlerFunc(infra.update).Path("/update")
	r.Methods("POST").HandlerFunc(infra.ddelete).Path("/delete")
	r.Methods("GET").HandlerFunc(infra.sselect).Path("/select")
	r.Methods("POST").HandlerFunc(infra.insert).Path("/insert")

	// just as sample, we put server here
	server := &http.Server{
		Addr:           ":8080",
		Handler:        c.Handler(r),
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	infra.server = server

	return nil
}

func (infra *Infrastructure) Depends() []interface{} {
	return []interface{}{
		infra.AddMiddleware,
	}
}

func (infra *Infrastructure) AddMiddleware(m Middleware) error {
	infra.mdwr = append(infra.mdwr, m)
	return nil
}

func (infra *Infrastructure) Close() error {
	return nil
}

///////////////// INFRA HANDLERS //////////////////////////////

func (infra *Infrastructure) update(writer http.ResponseWriter, request *http.Request) {
	infra.db.Update()
	writer.WriteHeader(http.StatusOK)
}

// ddelete just to not collide with delete keyword
func (infra *Infrastructure) ddelete(writer http.ResponseWriter, request *http.Request) {
	infra.db.Delete()
	writer.WriteHeader(http.StatusOK)
}

// sselect just to not collide with select keyword
func (infra *Infrastructure) sselect(writer http.ResponseWriter, request *http.Request) {
	infra.db.Select()
	writer.WriteHeader(http.StatusOK)
}
func (infra *Infrastructure) insert(writer http.ResponseWriter, request *http.Request) {
	infra.db.Insert()
	writer.WriteHeader(http.StatusOK)
}
