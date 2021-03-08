package endure

import "time"

// SetLogLevel option sets the log level in the Endure
func SetLogLevel(lvl Level) Options {
	return func(endure *Endure) {
		endure.loglevel = lvl
	}
}

// RetryOnFail if set to true, endure will try to stop and restart graph if one or more vertices are failed
func RetryOnFail(retry bool) Options {
	return func(endure *Endure) {
		endure.retry = retry
	}
}

// SetBackoff sets initial and maximum backoff interval for retry
func SetBackoff(initialInterval time.Duration, maxInterval time.Duration) Options {
	return func(endure *Endure) {
		endure.maxInterval = maxInterval
		endure.initialInterval = initialInterval
	}
}

// Visualize visualize current graph. Output: can be file or stdout
func Visualize(output Output, path string) Options {
	return func(endure *Endure) {
		endure.output = output
		if path != "" {
			endure.path = path
		}
	}
}

// GracefulShutdownTimeout sets the timeout to kill the vertices is one or more of them are frozen
func GracefulShutdownTimeout(to time.Duration) Options {
	return func(endure *Endure) {
		endure.stopTimeout = to
	}
}
