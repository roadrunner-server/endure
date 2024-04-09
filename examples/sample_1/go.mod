module samples

go 1.22.2

require (
	github.com/NYTimes/gziphandler v1.1.1
	github.com/gorilla/mux v1.8.0
	github.com/roadrunner-server/endure/v2 v2.3.1
	github.com/rs/cors v1.9.0
	go.etcd.io/bbolt v1.3.7
)

replace github.com/roadrunner-server/endure/v2 => ../../

require (
	github.com/roadrunner-server/errors v1.4.0 // indirect
	golang.org/x/sync v0.7.0 // indirect
	golang.org/x/sys v0.11.0 // indirect
)
