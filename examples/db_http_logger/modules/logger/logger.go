package logger

import (
	"fmt"

	"github.com/spiral/cascade"
)

type Logger struct {
}

type SuperLogger interface {
	SuperLogToStdOut(message string)
}

func (l *Logger) SuperLogToStdOut(message string) {
	// BOOM
	fmt.Println("logger says: " + message)
}

func (l *Logger) Init() error {
	return nil
}

func (l *Logger) Serve() chan error {
	errCh := make(chan error)
	return errCh
}

func (l *Logger) Stop() error {
	return nil
}

func (l *Logger) Provides() []interface{} {
	return []interface{}{
		l.LoggerInstance,
	}
}

func (l *Logger) LoggerInstance(name cascade.Named) (*Logger, error) {
	println(name.Name() + " invoke " + "logger")
	return l, nil
}