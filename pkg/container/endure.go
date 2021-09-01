package endure

import (
	"net/http"
	// pprof will be enabled in debug mode
	"net/http/pprof"

	"reflect"
	"sync"
	"time"

	"github.com/spiral/endure/pkg/fsm"
	"github.com/spiral/endure/pkg/graph"
	"github.com/spiral/endure/pkg/linked_list"
	"github.com/spiral/errors"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var order = 1

const (
	// InitializeMethodName is the method fn to invoke in transition map
	InitializeMethodName = "Initialize"
	// StartMethodName is the method fn to invoke in transition map
	StartMethodName = "Start"
	// ShutdownMethodName is the method fn to invoke in transition map
	ShutdownMethodName = "Shutdown"
)

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

// Endure struct represent main endure repr
type Endure struct {
	// Dependency graph
	graph *graph.Graph
	// DLL used as run list to run in order
	runList *linked_list.DoublyLinkedList
	// logger
	logger *zap.Logger
	// OPTIONS
	// retry on vertex fail
	retry           bool
	maxInterval     time.Duration
	initialInterval time.Duration
	stopTimeout     time.Duration

	depsOrder []string
	deps      map[string]interface{}
	disabled  map[string]bool

	// Graph visualizer
	// option to out to os.stdout or write data to file
	output Output
	// in case of file -> provide path to the file
	path string

	// internal loglevel in case if used internal logger. default -> Debug
	loglevel Level

	// Endure state machine
	fsm.FSM

	mutex *sync.RWMutex

	// result always points on healthy channel associated with vertex
	// since Endure structure has ALL method with pointer receiver, we do not need additional pointer to the sync.Map
	results sync.Map
	// main thread
	handleErrorCh chan *result
	userResultsCh chan *Result
}

// Options is the endure options
type Options func(endure *Endure)

/*
NewContainer returns empty endure container
Input parameters: logLevel
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
func NewContainer(logger *zap.Logger, options ...Options) (*Endure, error) {
	const op = errors.Op("new_container")
	c := &Endure{
		mutex:           &sync.RWMutex{},
		initialInterval: time.Second * 1,
		maxInterval:     time.Second * 60,
		results:         sync.Map{},
		stopTimeout:     time.Second * 10,
		loglevel:        DebugLevel,
		path:            "",
		// default empty -> no output
		output:    Empty,
		disabled:  make(map[string]bool),
		deps:      make(map[string]interface{}),
		depsOrder: make([]string, 0, 2),
	}

	// Transition map
	transitionMap := make(map[fsm.Event]reflect.Method)
	init, _ := reflect.TypeOf(c).MethodByName(InitializeMethodName)
	// event -> Initialize
	transitionMap[fsm.Initialize] = init

	serve, _ := reflect.TypeOf(c).MethodByName(StartMethodName)
	// event -> Start
	transitionMap[fsm.Start] = serve

	shutdown, _ := reflect.TypeOf(c).MethodByName(ShutdownMethodName)
	// event -> Stop
	transitionMap[fsm.Stop] = shutdown

	c.FSM = fsm.NewFSM(fsm.Uninitialized, transitionMap)

	c.graph = graph.NewGraph()
	c.runList = linked_list.NewDoublyLinkedList()

	// Main thread channels
	c.handleErrorCh = make(chan *result)
	c.userResultsCh = make(chan *Result)

	// append options
	for _, option := range options {
		option(c)
	}

	if logger == nil {
		log, err := c.internalLogger()
		if err != nil {
			return nil, errors.E(op, err)
		}
		c.logger = log
	} else {
		c.logger = logger
	}

	return c, nil
}

func (e *Endure) internalLogger() (*zap.Logger, error) {
	const op = errors.Op("endure_internal_logger")
	var lvl zap.AtomicLevel
	switch e.loglevel {
	case DebugLevel:
		lvl = zap.NewAtomicLevelAt(zap.DebugLevel)
		// start pprof
		profile()
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

	return logger, nil
}

func profile() {
	go func() {
		mux := http.NewServeMux()

		mux.HandleFunc("/debug/pprof/", pprof.Index)
		mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
		mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
		mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
		mux.HandleFunc("/debug/pprof/trace", pprof.Trace)

		srv := &http.Server{Handler: mux, Addr: "0.0.0.0:6061"}

		_ = srv.ListenAndServe()
	}()
}

// Register registers the dependencies in the Endure graph without invoking any methods
func (e *Endure) Register(vertex interface{}) error {
	const op = errors.Op("endure_register")
	t := reflect.TypeOf(vertex)
	vertexID := removePointerAsterisk(t.String())

	// if vertex disabled - skip
	if _, ok := e.disabled[vertexID]; ok {
		return nil
	}

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
	4. Provided type String fn
	We add 3 and 4 points to the Vertex
	*/
	err = e.addProviders(vertexID, vertex)
	if err != nil {
		return errors.E(op, errors.Providers, err)
	}
	e.logger.Debug("registering type", zap.String("type", t.String()))

	// if deps present in the map - skip
	if _, ok := e.deps[vertexID]; ok {
		return nil
	}

	// save all vertices on the initial stage
	e.deps[vertexID] = vertex
	e.depsOrder = append(e.depsOrder, vertexID)

	return nil
}

// RegisterAll is the helper for the register to register more than one structure in the endure
func (e *Endure) RegisterAll(plugins ...interface{}) error {
	const op = errors.Op("endure_register_all")
	for _, plugin := range plugins {
		err := e.Register(plugin)
		if err != nil {
			return errors.E(op, err)
		}
	}

	return nil
}

// Init container and all service edges.
func (e *Endure) Init() error {
	_, err := e.Transition(fsm.Initialize, e)
	if err != nil {
		return err
	}
	return nil
}

// Serve starts serving the graph
// This is the initial serveInternal, if error produced immediately in the initial serveInternal, endure will traverse deps back, call internal_stop and exit
func (e *Endure) Serve() (<-chan *Result, error) {
	data, err := e.Transition(fsm.Start, e)
	if err != nil {
		return nil, err
	}
	// god save this construction
	return data.(<-chan *Result), nil
}

// Stop stops the execution and call Stop on every vertex
func (e *Endure) Stop() error {
	_, err := e.Transition(fsm.Stop, e)
	if err != nil {
		return err
	}
	return nil
}

// Initialize used to add edges between vertices, sort graph topologically
// Do not change this method fn, sync with constants in the beginning of this file
func (e *Endure) Initialize() error {
	const op = errors.Op("endure_initialize")
	// START used to restart Initialize when disabled vertex found
	// TODO temporary solution
START:
	// traverse the graph
	err := e.addEdges()
	if err != nil {
		return errors.E(op, errors.Init, err)
	}

	// if failed - continue, just send warning to a user
	// visualize is not critical
	if e.output != Empty {
		err = e.Visualize(e.graph.Vertices)
		if err != nil {
			e.logger.Warn("failed to visualize the graph", zap.Error(err))
		}
	}

	// we should build internal_init list in the reverse order
	sorted, err := graph.TopologicalSort(e.graph.Vertices)
	if err != nil {
		e.logger.Error("error sorting the graph", zap.Error(err))
		return errors.E(op, errors.Init, err)
	}

	// >= because disabled also contains vertex provided values
	if len(e.disabled) >= len(e.deps) {
		e.logger.Error("all vertices are disabled: graph should contain at least 1 active vertex, possibly all vertices was disabled because of ROOT vertex failure", zap.Any("disabled", e.disabled))
		return errors.E(op, errors.Init, errors.Errorf("graph should contain at least 1 active vertex, possibly all vertices was disabled because of ROOT vertex failure"))
	}

	if len(sorted) == 0 && len(e.disabled) != 0 {
		e.logger.Error("initial graph should contain at least 1 vertex, possibly you forget to invoke Registers?")
		return errors.E(op, errors.Init, errors.Errorf("graph should contain at least 1 vertex, possibly you forget to invoke registers"))
	}

	e.runList = linked_list.NewDoublyLinkedList()
	for i := len(sorted) - 1; i >= 0; i-- {
		e.runList.Push(sorted[i])
	}

	head := e.runList.Head
	headCopy := head
	for headCopy != nil {
		// check for disabled, because that can be interface
		if _, ok := e.disabled[headCopy.Vertex.ID]; ok {
			err = e.removeVertex(headCopy)
			if err != nil {
				return errors.E(op, err)
			}
			// start from the clear state excluding the disabled vertex
			goto START
		}
		headCopy.Vertex.SetState(fsm.Initializing)
		err = e.internalInit(headCopy.Vertex)
		if err != nil {
			// remove head
			if errors.Is(errors.Disabled, err) {
				err = e.removeVertex(headCopy)
				if err != nil {
					return errors.E(op, err)
				}

				// start from the clear state excluding the disabled vertex
				goto START
			}

			headCopy.Vertex.SetState(fsm.Error)
			e.logger.Error("error during the internal_init", zap.Error(err))
			return errors.E(op, errors.Init, err)
		}
		headCopy.Vertex.SetState(fsm.Initialized)
		headCopy = headCopy.Next
	}

	return nil
}

// Start used to start serving vertices
// Do not change this method fn, sync with constants in the beginning of this file
func (e *Endure) Start() (<-chan *Result, error) {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	const op = errors.Op("endure_start")
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
		nCopy.Vertex.SetState(fsm.Starting)
		err := e.serveInternal(nCopy)
		if err != nil {
			nCopy.Vertex.SetState(fsm.Error)
			e.traverseBackStop(nCopy)
			return nil, errors.E(op, errors.Serve, err)
		}
		nCopy.Vertex.SetState(fsm.Started)
		nCopy = nCopy.Next
	}
	// all vertices disabled
	if !atLeastOne {
		return nil, errors.E(op, errors.Disabled, errors.Str("all vertices disabled, nothing to serveInternal"))
	}
	return e.userResultsCh, nil
}

// Shutdown used to shutdown the Endure
// Do not change this method fn, sync with constants in the beginning of this file
func (e *Endure) Shutdown() error {
	e.logger.Info("exiting from the Endure")
	n := e.runList.Head
	return e.shutdown(n, true)
}

func (e *Endure) removeVertex(head *linked_list.DllNode) error {
	const op = errors.Op("endure_disable")
	e.logger.Debug("found disabled vertex", zap.String("id", head.Vertex.ID))
	// add vertex to the map with disabled vertices
	for providesID := range head.Vertex.Provides {
		e.disabled[providesID] = true
	}
	e.disabled[head.Vertex.ID] = true
	// reset run list
	e.runList.Reset()
	// reset graph
	e.graph.ClearState()

	// re-register all deps, excluding disabled preserving initial order
	for i := 0; i < len(e.depsOrder); i++ {
		err := e.Register(e.deps[e.depsOrder[i]])
		if err != nil {
			return errors.E(op, err)
		}
	}

	return nil
}
