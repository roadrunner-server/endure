package endure

import "time"

// GracefulShutdownTimeout sets the timeout to kill the vertices is one or more of them are frozen
func GracefulShutdownTimeout(to time.Duration) Options {
	return func(endure *Endure) {
		endure.stopTimeout = to
	}
}

func Visualize() Options {
	return func(endure *Endure) {
		endure.visualize = true
	}
}

func EnableProfiler() Options {
	return func(endure *Endure) {
		endure.profiler = true
	}
}
