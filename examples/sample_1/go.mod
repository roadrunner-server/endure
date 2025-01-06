module samples

go 1.23

toolchain go1.23.0

require (
	github.com/NYTimes/gziphandler v1.1.1
	github.com/gorilla/mux v1.8.1
	github.com/roadrunner-server/endure/v2 v2.3.1
	github.com/rs/cors v1.11.0
	go.etcd.io/bbolt v1.3.10
)

replace github.com/roadrunner-server/endure/v2 => ../../

require (
	github.com/fatih/color v1.18.0 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/roadrunner-server/errors v1.4.1 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	go.uber.org/zap v1.27.0 // indirect
	golang.org/x/sync v0.8.0 // indirect
	golang.org/x/sys v0.29.0 // indirect
)
