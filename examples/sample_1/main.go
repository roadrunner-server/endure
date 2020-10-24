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
	container, err := endure.NewContainer(endure.DebugLevel, endure.RetryOnFail(true), endure.PrintGraph(true))
	if err != nil {
		panic(err)
	}

	err = container.Register(&http.Http{})
	if err != nil {
		panic(err)
	}
	err = container.Register(&db.DB{})
	if err != nil {
		panic(err)
	}
	err = container.Register(&logger.Logger{})
	if err != nil {
		panic(err)
	}
	err = container.Register(&gzip.Gzip{})
	if err != nil {
		panic(err)
	}
	err = container.Register(&headers.Headers{})
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
