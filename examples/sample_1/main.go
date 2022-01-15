package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/roadrunner-server/endure/examples/db_http_logger/modules/db"
	"github.com/roadrunner-server/endure/examples/db_http_logger/modules/gzip"
	"github.com/roadrunner-server/endure/examples/db_http_logger/modules/headers"
	"github.com/roadrunner-server/endure/examples/db_http_logger/modules/http"
	"github.com/roadrunner-server/endure/examples/db_http_logger/modules/logger"
	endure "github.com/roadrunner-server/endure/pkg/container"
)

func main() {
	// no external logger
	container, err := endure.NewContainer(nil, endure.RetryOnFail(true), endure.Visualize(endure.StdOut, ""), endure.SetLogLevel(endure.DebugLevel))
	if err != nil {
		panic(err)
	}

	err = container.RegisterAll(
		&http.Http{},
		&db.DB{},
		&logger.Logger{},
		&gzip.Gzip{},
		&headers.Headers{},
	)

	if err != nil {
		panic(err)
	}

	err = container.Init()
	if err != nil {
		panic(err)
	}

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
