package http

import (
	"context"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/rs/cors"
	"github.com/spiral/cascade/samples/db_http_logger/modules/db"
	"github.com/spiral/cascade/samples/db_http_logger/modules/logger"
)

type Infrastructure struct {
	client http.Client
	server *http.Server
	mdwr   []middleware
	db     db.Repository
	logger logger.SuperLogger
}

// http middleware type.
type middleware func(f http.HandlerFunc) http.HandlerFunc

func (infra *Infrastructure) Init(db db.Repository, logger logger.SuperLogger) error {
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
	errCh := make(chan error, 1)

	f := infra.server.Handler.ServeHTTP

	// chain middlewares
	for i := 0; i < len(infra.mdwr); i++ {
		infra.mdwr[i](f)
	}

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

func (infra *Infrastructure) AddMiddleware(m middleware) {
	infra.mdwr = append(infra.mdwr, m)
}

func (infra *Infrastructure) Close() error {
	return nil
}

///////////////// INFRA HANDLERS //////////////////////////////

func (infra *Infrastructure) update(writer http.ResponseWriter, request *http.Request) {
	infra.db.Update()
}

// ddelete just to not collide with delete keyword
func (infra *Infrastructure) ddelete(writer http.ResponseWriter, request *http.Request) {
	infra.db.Delete()
}

// sselect just to not collide with select keyword
func (infra *Infrastructure) sselect(writer http.ResponseWriter, request *http.Request) {
	infra.db.Select()
}
func (infra *Infrastructure) insert(writer http.ResponseWriter, request *http.Request) {
	infra.db.Insert()
}
