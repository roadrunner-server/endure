package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/spiral/endure"
	"github.com/spiral/endure/examples/db_http_logger/modules/db"
	"github.com/spiral/endure/examples/db_http_logger/modules/gzip"
	"github.com/spiral/endure/examples/db_http_logger/modules/headers"
	"github.com/spiral/endure/examples/db_http_logger/modules/http"
	"github.com/spiral/endure/examples/db_http_logger/modules/logger"
)

func main() {
	// no external logger
	container, err := endure.NewContainer(nil, endure.RetryOnFail(true), endure.Visualize(true), endure.SetLogLevel(endure.DebugLevel))
	if err != nil {
		panic(err)
	}

	err = InitApp(
		container,
		&http.Http{},
		&db.DB{},
		&logger.Logger{},
		&gzip.Gzip{},
		&headers.Headers{},
	)

	errCh, err := container.Serve()
	if err != nil {
		panic(err)
	}

	// stop by CTRL+C
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGKILL, syscall.SIGINT)

	for {
		select {
		case e := <-errCh:
			println(e.Error.Error())
			er := container.Stop()
			if er != nil {
				panic(er)
			}
			return
		case <-c:
			er := container.Stop()
			if er != nil {
				panic(er)
			}
			return
		}
	}
}

// InitApp with a list of provided services.
func InitApp(container endure.Container, service ...interface{}) error {
	for _, svc := range service {
		err := container.Register(svc)
		if err != nil {
			return err
		}
	}

	err := container.Init()
	if err != nil {
		return err
	}

	return nil
}
