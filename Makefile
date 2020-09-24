test:
	go test -v -race ./tests/backoff
	go test -v -race ./tests/happyScenario
	go test -v -race ./tests/interfaces
	go test -v -race ./tests/issues
	go test -v -race ./tests/stress
