module samples

go 1.27

require (
	github.com/NYTimes/gziphandler v1.1.1
	github.com/gorilla/mux v1.8.1
	github.com/roadrunner-server/endure/v2 v2.3.1
	github.com/rs/cors v1.11.0
	go.etcd.io/bbolt v1.5.0
)

replace github.com/roadrunner-server/endure/v2 => ../../

require (
	github.com/fatih/color v1.19.0 // indirect
	github.com/mattn/go-colorable v0.1.15 // indirect
	github.com/mattn/go-isatty v0.0.24 // indirect
	github.com/roadrunner-server/errors v1.5.0 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	go.uber.org/zap v1.28.0 // indirect
	golang.org/x/sys v0.47.0 // indirect
)
