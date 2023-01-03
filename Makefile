test:
	go test -v -race -tags=debug ./tests/init
	go test -v -race -tags=debug ./tests/happy_scenarios
	go test -v -race -tags=debug ./tests/interfaces
	go test -v -race -tags=debug ./tests/issues
	go test -v -race -tags=debug ./tests/stress
	go test -v -race -tags=debug ./tests/disabled_vertices
