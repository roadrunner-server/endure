test:
	go test -v -race ./tests/backoff
	go test -v -race ./tests/happy_scenarios
	go test -v -race ./tests/interfaces
	go test -v -race ./tests/issues
	go test -v -race ./tests/stress
	go test -v -race ./tests/disabled_vertices
