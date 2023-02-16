test:
	go test -v -race -tags=debug ./tests/general/test1
	go test -v -race -tags=debug ./tests/general/test2
	go test -v -race -tags=debug ./tests/general/test3
	go test -v -race -tags=debug ./tests/general/test4
	go test -v -race -tags=debug ./tests/general/test5
	go test -v -race -tags=debug ./tests/init
	go test -v -race -tags=debug ./tests/happy_scenarios
	go test -v -race -tags=debug ./tests/interfaces
	go test -v -race -tags=debug ./tests/issues
	go test -v -race -tags=debug ./tests/stress
	go test -v -race -tags=debug ./tests/disabled_vertices
