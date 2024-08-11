module samples

go 1.22.5

require (
	github.com/NYTimes/gziphandler v1.1.1
	github.com/gorilla/mux v1.8.1
	github.com/roadrunner-server/endure/v2 v2.3.1
	github.com/rs/cors v1.11.0
	go.etcd.io/bbolt v1.3.10
)

replace github.com/roadrunner-server/endure/v2 => ../../

require (
	github.com/roadrunner-server/errors v1.4.1 // indirect
	golang.org/x/sync v0.8.0 // indirect
	golang.org/x/sys v0.22.0 // indirect
)
