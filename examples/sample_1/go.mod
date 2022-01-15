module github.com/roadrunner-server/endure/examples/db_http_logger

go 1.15

require (
	github.com/NYTimes/gziphandler v1.1.1
	github.com/gorilla/mux v1.7.4
	github.com/rs/cors v1.7.0
	github.com/roadrunner-server/endure v1.0.0-beta17
	go.etcd.io/bbolt v1.3.5
)

replace github.com/roadrunner-server/endure => ../../
