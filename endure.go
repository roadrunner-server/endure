package endure

import (
	"errors"
	"net/http"

	// pprof will be enabled in debug mode
	_ "net/http/pprof"

	"reflect"
	"sync"
	"time"

	"github.com/spiral/endure/structures"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var order int = 1

// A Level is a logging priority. Higher levels are more important.
type Level int8

const (
	// DebugLevel logs are typically voluminous, and are usually disabled in
	// production.
	DebugLevel Level = iota - 1
	// InfoLevel is the default logging priority.
	InfoLevel
	// WarnLevel logs are more important than Info, but don't need individual
	// human review.
	WarnLevel
	// ErrorLevel logs are high-priority. If an application is running smoothly,
	// it shouldn't generate any error-level logs.
	ErrorLevel
	// DPanicLevel logs are particularly important errors. In development the
	// logger panics after writing the message.
	DPanicLevel
	// PanicLevel logs a message, then panics.
	PanicLevel
	// FatalLevel logs a message, then calls os.Exit(1).
	FatalLevel
)

type Endure struct {
	// Dependency graph
	graph *structures.Graph
	// DLL used as run list to run in order
	runList *structures.DoublyLinkedList
	// logger
	logger *zap.Logger
	// OPTIONS
	// retry on vertex fail
	retry           bool
	maxInterval     time.Duration
	initialInterval time.Duration

	mutex *sync.RWMutex

	// result always points on healthy channel associated with vertex
	results map[string]*result

	// main thread
	handleErrorCh chan *result
	userResultsCh chan *Result
}

type Options func(endure *Endure)

/* Input parameters: logLevel
   -1 is the most informative level - DebugLevel --> also turns on pprof endpoint
   0 - InfoLevel defines info log level.
   1 -
   2 - WarnLevel defines warn log level.
   3 - ErrorLevel defines error log level.
   4 - FatalLevel defines fatal log level.
   5 - PanicLevel defines panic log level.
   6 - NoLevel defines an absent log level.
   7 - Disabled disables the logger.
   see the endure.Level
*/
func NewContainer(logLevel Level, options ...Options) (*Endure, error) {
	c := &Endure{
		mutex:           &sync.RWMutex{},
		initialInterval: time.Second * 1,
		maxInterval:     time.Second * 60,
	}

	var lvl zap.AtomicLevel
	switch logLevel {
	case DebugLevel:
		lvl = zap.NewAtomicLevelAt(zap.DebugLevel)
		// start pprof
		pprof()
	case InfoLevel:
		lvl = zap.NewAtomicLevelAt(zap.InfoLevel)
	case WarnLevel:
		lvl = zap.NewAtomicLevelAt(zap.WarnLevel)
	case ErrorLevel:
		lvl = zap.NewAtomicLevelAt(zap.ErrorLevel)
	case FatalLevel:
		lvl = zap.NewAtomicLevelAt(zap.FatalLevel)
	case PanicLevel:
		lvl = zap.NewAtomicLevelAt(zap.PanicLevel)
	case DPanicLevel:
		lvl = zap.NewAtomicLevelAt(zap.DPanicLevel)
	default:
		lvl = zap.NewAtomicLevelAt(zap.InfoLevel)
	}

	cfg := zap.Config{
		Level:    lvl,
		Encoding: "console",
		EncoderConfig: zapcore.EncoderConfig{
			MessageKey:    "message",
			LevelKey:      "level",
			TimeKey:       "time",
			CallerKey:     "caller",
			StacktraceKey: "stack",
			EncodeLevel:   zapcore.CapitalLevelEncoder,
			EncodeTime:    zapcore.ISO8601TimeEncoder,
			EncodeCaller:  zapcore.ShortCallerEncoder,
		},
		OutputPaths:      []string{"stderr"},
		ErrorOutputPaths: []string{"stderr"},
	}

	logger, err := cfg.Build(zap.AddCaller())
	if err != nil {
		return nil, err
	}
	c.logger = logger

	c.graph = structures.NewGraph()
	c.runList = structures.NewDoublyLinkedList()
	c.logger = logger
	c.results = make(map[string]*result)

	// Main thread channels
	c.handleErrorCh = make(chan *result)
	c.userResultsCh = make(chan *Result)

	// append options
	for _, option := range options {
		option(c)
	}

	return c, nil
}

func pprof() {
	go func() {
		_ = http.ListenAndServe("0.0.0.0:6061", nil)
	}()
}

func RetryOnFail(set bool) Options {
	return func(endure *Endure) {
		endure.retry = set
	}
}

func SetBackoffTimes(initialInterval time.Duration, maxInterval time.Duration) Options {
	return func(endure *Endure) {
		endure.maxInterval = maxInterval
		endure.initialInterval = initialInterval
	}
}

// Depender depends the dependencies
// name is a name of the dependency, for example - S2
// vertex is a value -> pointer to the structure
func (e *Endure) Register(vertex interface{}) error {
	t := reflect.TypeOf(vertex)
	vertexID := removePointerAsterisk(t.String())

	if t.Kind() != reflect.Ptr {
		return errors.New("you should pass pointer to the structure instead of value")
	}

	ok := t.Implements(reflect.TypeOf((*Service)(nil)).Elem())
	if !ok {
		return errTypeNotImplementError
	}

	/* Depender the type
	Information we know at this step is:
	1. vertexID
	2. Vertex structure value (interface)
	And we fill vertex with this information
	*/
	err := e.register(vertexID, vertex, order)
	if err != nil {
		return err
	}
	order++
	/* Add the types, which (if) current vertex provides
	Information we know at this step is:
	1. vertexID
	2. Vertex structure value (interface)
	3. Provided type
	4. Provided type String name
	We add 3 and 4 points to the Vertex
	*/
	err = e.addProviders(vertexID, vertex)
	if err != nil {
		return err
	}
	e.logger.Debug("registering type", zap.String("type", t.String()))

	return nil
}

// Init container and all service edges.
func (e *Endure) Init() error {
	// traverse the graph
	if err := e.addEdges(); err != nil {
		return err
	}

	// we should build init list in the reverse order
	sorted, err := structures.TopologicalSort(e.graph.Vertices)
	if err != nil {
		e.logger.Error("error sorting the graph", zap.Error(err))
		return err
	}

	if len(sorted) == 0 {
		e.logger.Error("initial graph should contain at least 1 vertex, possibly you forget to invoke Registers?")
		return errors.New("graph should contain at least 1 vertex, possibly you forget to invoke registers")
	}

	e.runList = structures.NewDoublyLinkedList()
	for i := 0; i <= len(sorted)-1; i++ {
		e.runList.Push(sorted[i])
	}

	head := e.runList.Head
	headCopy := head
	for headCopy != nil {
		err := e.init(headCopy.Vertex)
		if err != nil {
			e.logger.Error("error during the init", zap.Error(err))
			return err
		}
		headCopy = headCopy.Next
	}

	return nil
}

func (e *Endure) Serve() (<-chan *Result, error) {
	e.startMainThread()

	// call configure
	nCopy := e.runList.Head
	for nCopy != nil {
		err := e.configure(nCopy)
		if err != nil {
			e.logger.Error("backoff failed", zap.String("vertex id", nCopy.Vertex.ID), zap.Error(err))
			return nil, err
		}

		nCopy = nCopy.Next
	}

	nCopy = e.runList.Head
	for nCopy != nil {
		err := e.serve(nCopy)
		if err != nil {
			return nil, err
		}
		nCopy = nCopy.Next
	}
	return e.userResultsCh, nil
}

func (e *Endure) Stop() error {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	e.logger.Info("exiting from the Endure")
	n := e.runList.Head
	e.shutdown(n)
	return nil
}
