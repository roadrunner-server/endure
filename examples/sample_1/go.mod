module samples

go 1.19

require (
	github.com/NYTimes/gziphandler v1.1.1
	github.com/gorilla/mux v1.8.0
	github.com/roadrunner-server/endure/v2 v2.0.0
	github.com/rs/cors v1.8.3
	go.etcd.io/bbolt v1.3.6
)

replace github.com/roadrunner-server/endure/v2 => ../../

require (
	github.com/roadrunner-server/errors v1.2.0 // indirect
	golang.org/x/exp v0.0.0-20221230185412-738e83a70c30 // indirect
	golang.org/x/sys v0.3.0 // indirect
)
