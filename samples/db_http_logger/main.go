package main

import (
	"github.com/spiral/cascade"
	"github.com/spiral/cascade/samples/db_http_logger/db"
	"github.com/spiral/cascade/samples/db_http_logger/gzip_plugin"
	"github.com/spiral/cascade/samples/db_http_logger/http"
	"github.com/spiral/cascade/samples/db_http_logger/logger"
)

func main() {
	container, err := cascade.NewContainer(cascade.DebugLevel, cascade.RetryOnFail(true))
	if err != nil {
		panic(err)
	}

	err = container.Register(&http.Infrastructure{})
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

	err = container.Register(&gzip_plugin.GzipPlugin{})
	if err != nil {
		panic(err)
	}

	err = container.Init()
	if err != nil {
		panic(err)
	}

	err, errCh := container.Serve()
	if err != nil {
		panic(err)
	}

	for {
		select {
		case e := <-errCh:
			println(e.Error.Err.Error())
		}
	}
}
