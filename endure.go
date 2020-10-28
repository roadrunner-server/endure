package endure

import (
	"net/http"

	// pprof will be enabled in debug mode
	_ "net/http/pprof"

	"reflect"
	"sync"
	"time"

	"github.com/spiral/endure/structures"
	"github.com/spiral/errors"
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
	// option to visualize resulted (before init) graph
	visualize bool

	mutex *sync.RWMutex

	// result always points on healthy channel associated with vertex
	// since Endure structure has ALL method with pointer receiver, we do not need additional pointer to the sync.Map
	results sync.Map
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
	const op = errors.Op("NewContainer")
	c := &Endure{
		mutex:           &sync.RWMutex{},
		initialInterval: time.Second * 1,
		maxInterval:     time.Second * 60,
		results:         sync.Map{},
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
		return nil, errors.E(op, errors.Logger, err)
	}
	c.logger = logger

	c.graph = structures.NewGraph()
	c.runList = structures.NewDoublyLinkedList()
	c.logger = logger

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

func RetryOnFail(retry bool) Options {
	return func(endure *Endure) {
		endure.retry = retry
	}
}

func SetBackoffTimes(initialInterval time.Duration, maxInterval time.Duration) Options {
	return func(endure *Endure) {
		endure.maxInterval = maxInterval
		endure.initialInterval = initialInterval
	}
}

func Visualize(print bool) Options {
	return func(endure *Endure) {
		endure.visualize = print
	}
}

// Register registers the dependencies in the Endure graph without invoking any methods
func (e *Endure) Register(vertex interface{}) error {
	const op = errors.Op("Register")
	t := reflect.TypeOf(vertex)
	vertexID := removePointerAsterisk(t.String())

	if t.Kind() != reflect.Ptr {
		return errors.E(op, errors.Register, errors.Errorf("you should pass pointer to the structure instead of value"))
	}

	/* Collector the type
	Information we know at this step is:
	1. vertexID
	2. Vertex structure value (interface)
	And we fill vertex with this information
	*/
	err := e.register(vertexID, vertex, order)
	if err != nil {
		return errors.E(op, errors.Register, err)
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
		return errors.E(op, errors.Providers, err)
	}
	e.logger.Debug("registering type", zap.String("type", t.String()))

	return nil
}

// Init container and all service edges.
func (e *Endure) Init() error {
	const op = errors.Op("Init")
	// traverse the graph
	err := e.addEdges()
	if err != nil {
		return errors.E(op, errors.Init, err)
	}

	// if failed - continue, just send warning to a user
	// visualize is not critical
	if e.visualize {
		err = structures.Visualize(e.graph.Vertices)
		if err != nil {
			e.logger.Warn("failed to visualize the graph", zap.Error(err))
		}
	}

	// we should build init list in the reverse order
	sorted, err := structures.TopologicalSort(e.graph.Vertices)
	if err != nil {
		e.logger.Error("error sorting the graph", zap.Error(err))
		return errors.E(op, errors.Init, err)
	}

	if len(sorted) == 0 {
		e.logger.Error("initial graph should contain at least 1 vertex, possibly you forget to invoke Registers?")
		return errors.E(op, errors.Init, errors.Errorf("graph should contain at least 1 vertex, possibly you forget to invoke registers"))
	}

	e.runList = structures.NewDoublyLinkedList()
	for i := len(sorted) - 1; i >= 0; i-- {
		e.runList.Push(sorted[i])
	}

	head := e.runList.Head
	headCopy := head
	for headCopy != nil {
		err = e.init(headCopy.Vertex)
		if err != nil {
			e.logger.Error("error during the init", zap.Error(err))
			return errors.E(op, errors.Init, err)
		}
		headCopy = headCopy.Next
	}

	return nil
}

// Serve starts serving the graph
// This is the initial serve, if error produced immediately in the initial serve, endure will traverse deps back, call stop and exit
func (e *Endure) Serve() (<-chan *Result, error) {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	const op = errors.Op("Serve")
	e.startMainThread()

	// simple check that we have at least one vertex in the graph to Serve
	atLeastOne := false

	nCopy := e.runList.Head
	for nCopy != nil {
		if nCopy.Vertex.IsDisabled {
			nCopy = nCopy.Next
			continue
		}
		atLeastOne = true
		err := e.serve(nCopy)
		if err != nil {
			e.traverseBackStop(nCopy)
			return nil, errors.E(op, errors.Serve, err)
		}
		nCopy = nCopy.Next
	}
	// all vertices disabled
	if atLeastOne == false {
		return nil, errors.E(op, errors.Disabled, errors.Str("all vertices disabled, nothing to serve"))
	}
	return e.userResultsCh, nil
}

// Stop stops the execution and call Stop on every vertex
func (e *Endure) Stop() error {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	e.logger.Info("exiting from the Endure")
	n := e.runList.Head
	e.shutdown(n)
	return nil
}
