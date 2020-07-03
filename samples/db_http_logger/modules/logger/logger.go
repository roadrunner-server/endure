package logger

import (
	"errors"
	"fmt"
	"time"
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
	go func() {
		time.Sleep(time.Second * 1)
		errCh <- errors.New("test error from logger")
	}()
	return errCh
}

func (l *Logger) Stop() error {
	return nil
}
