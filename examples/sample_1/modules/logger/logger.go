package logger

import (
	"fmt"

	"github.com/roadrunner-server/endure/v2/dep"
)

type Named interface {
	// Name return user friendly name of the plugin
	Name() string
}

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

func (l *Logger) Provides() []*dep.Out {
	return []*dep.Out{
		dep.Bind((*SuperLogger)(nil), l.LoggerInstance),
	}
}

func (l *Logger) LoggerInstance() (*Logger, error) {
	println("logger invoked")
	return l, nil
}
