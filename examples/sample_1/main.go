package main

import (
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"samples/modules/db"
	"samples/modules/gzip"
	"samples/modules/headers"
	"samples/modules/http"
	"samples/modules/logger"

	"github.com/roadrunner-server/endure/v2"
)

func main() {
	// no external logger
	container := endure.New(slog.LevelDebug)

	// register your plugins here (references to the structures (aka implementations)
	err := container.RegisterAll(
		&http.Http{},
		&db.DB{},
		&logger.Logger{},
		&gzip.Gzip{},
		&headers.Headers{},
	)

	if err != nil {
		panic(err)
	}

	// Init the graph
	err = container.Init()
	if err != nil {
		panic(err)
	}

	// Start the graph
	errCh, err := container.Serve()
	if err != nil {
		panic(err)
	}

	// stop by CTRL+C
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

	for {
		select {
		// here is the error channel
		// endure will automatically stop all deps if receive an error
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
