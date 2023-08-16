package endure

import (
	"log/slog"
	"time"
)

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

// LogHandler defines the logger handler to create the slog.Logger
//
// For example:
//
//	container = endure.New(slog.LevelInfo, LogHandler(slog.NewTextHandler(
//	 os.Stderr, &slog.HandlerOptions{Level: slog.LevelInfo})))
func LogHandler(handler slog.Handler) Options {
	return func(endure *Endure) {
		endure.log = slog.New(handler)
	}
}
