module github.com/spiral/cascade/examples/db_http_logger

go 1.14

require (
	github.com/NYTimes/gziphandler v1.1.1
	github.com/gorilla/mux v1.7.4
	github.com/rs/cors v1.7.0
	github.com/spiral/cascade v1.0.0-beta4
	go.etcd.io/bbolt v1.3.5
)

replace (
	github.com/spiral/cascade v1.0.0-beta4 => "/home/valery/Projects/opensource/spiral/cascade"
)