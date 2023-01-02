package endure

import "time"

// SetBackoff sets initial and maximum backoff interval for retry
func SetBackoff(initialInterval time.Duration, maxInterval time.Duration) Options {
	return func(endure *Endure) {
		endure.maxInterval = maxInterval
		endure.initialInterval = initialInterval
	}
}

// GracefulShutdownTimeout sets the timeout to kill the vertices is one or more of them are frozen
func GracefulShutdownTimeout(to time.Duration) Options {
	return func(endure *Endure) {
		endure.stopTimeout = to
	}
}
