module samples

go 1.20

require (
	github.com/NYTimes/gziphandler v1.1.1
	github.com/gorilla/mux v1.8.0
	github.com/roadrunner-server/endure/v2 v2.0.1
	github.com/rs/cors v1.8.3
	go.etcd.io/bbolt v1.3.7
)

replace github.com/roadrunner-server/endure/v2 => ../../

require (
	github.com/roadrunner-server/errors v1.2.0 // indirect
	golang.org/x/exp v0.0.0-20230203172020-98cc5a0785f9 // indirect
	golang.org/x/sys v0.4.0 // indirect
)
