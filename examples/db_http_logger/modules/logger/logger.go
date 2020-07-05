package logger

import (
	"fmt"
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
