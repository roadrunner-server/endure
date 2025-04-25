package endure

import (
	"log/slog"
	"net/http"
	// pprof will be enabled in debug mode
	"net/http/pprof"
	"reflect"
	"sync"
	"time"

	"github.com/roadrunner-server/endure/v2/graph"
	"github.com/roadrunner-server/endure/v2/logger"
	"github.com/roadrunner-server/endure/v2/registar"
	"github.com/roadrunner-server/errors"
	"go.uber.org/zap"
)

// Endure struct represent main endure repr
type Endure struct {
	/// NEW
	registar *registar.Registar
	///
	mu sync.RWMutex
	// Dependency graph
	graph *graph.Graph
	// log
	log         *zap.Logger
	stopTimeout time.Duration
	profiler    bool
	visualize   bool

	// main thread
	handleErrorCh chan *result
	userResultsCh chan *Result
}

// Options is the 'endure' options
type Options func(endure *Endure)

// New returns empty endure container
func New(level slog.Leveler, options ...Options) *Endure {
	if level == nil {
		level = slog.LevelDebug
	}

	// error handling is omitted because we are sure that the logger will be created
	zlog, _ := logger.BuildLogger(level)

	c := &Endure{
		registar:    registar.New(),
		graph:       graph.New(),
		mu:          sync.RWMutex{},
		stopTimeout: time.Second * 30,
		log:         zlog.Named("endure"),
	}

	// Main thread channels
	c.handleErrorCh = make(chan *result)
	c.userResultsCh = make(chan *Result)

	// append options
	for _, option := range options {
		option(c)
	}

	// start profiler server
	if c.profiler {
		profile()
	}

	return c
}

// Register registers the dependencies in the Endure graph without invoking any methods
func (e *Endure) Register(vertex any) error {
	const op = errors.Op("endure_register")
	e.mu.Lock()
	defer e.mu.Unlock()

	t := reflect.TypeOf(vertex)

	// t.Kind() - ptr
	// t.Elem().Kind() - Struct
	if t.Kind() != reflect.Ptr {
		return errors.E(op, errors.Register, errors.Errorf("you should pass pointer to the structure instead of value"))
	}

	/* Collector the type
	Information we know at this step is:
	1. vertexID
	2. Vertex structure value (interface)
	And we fill vertex with this information
	*/

	if e.graph.HasVertex(vertex) {
		e.log.Warn("already registered", zap.Error(errors.Errorf("plugin `%s` is already registered", t.String())))
		return nil
	}

	weight := uint(1)
	if val, ok := vertex.(Weighted); ok {
		weight = val.Weight()
		e.log.Debug(
			"weight added",
			zap.String("type", reflect.TypeOf(vertex).Elem().String()),
			zap.String("kind", reflect.TypeOf(vertex).Elem().Kind().String()),
			zap.Uint64("value", uint64(weight)),
		)
	}

	// push the vertex
	e.graph.AddVertex(vertex, weight)
	// add the dependency for the resolver
	e.registar.Insert(vertex, reflect.TypeOf(vertex), "", weight)

	e.log.Debug(
		"type registered",
		zap.String("type", reflect.TypeOf(vertex).Elem().String()),
		zap.String("kind", reflect.TypeOf(vertex).Elem().Kind().String()),
		zap.String("method", "plugin"),
	)

	/*
		Add the types, which (if) current vertex provides
		Information we know at this step is:
		1. vertexID
		2. Vertex structure value (interface)
		3. Provided type
		4. Provided type String fn
		We add 3 and 4 points to the Vertex
	*/
	if val, ok := vertex.(Provider); ok {
		// get types
		outDeps := val.Provides()

		// iter
		for i := range outDeps {
			e.registar.Insert(vertex, outDeps[i].Type, outDeps[i].Method, weight)
			e.log.Debug(
				"provided type registered",
				zap.String("type", outDeps[i].Type.String()),
				zap.String("kind", outDeps[i].Type.Kind().String()),
				zap.String("method", outDeps[i].Method),
			)
		}
	}

	return nil
}

// RegisterAll is the helper for the register to register more than one structure in the endure
func (e *Endure) RegisterAll(plugins ...any) error {
	const op = errors.Op("endure_register_all")
	for _, plugin := range plugins {
		err := e.Register(plugin)
		if err != nil {
			return errors.E(op, err)
		}
	}

	return nil
}

func (e *Endure) Init() error {
	e.mu.Lock()
	defer e.mu.Unlock()

	const op = errors.Op("endure_initialize")

	if len(e.graph.Vertices()) == 0 {
		return errors.E(op, errors.Str("no plugins registered"))
	}

	// traverse the graph
	err := e.resolveEdges()
	if err != nil {
		return errors.E(op, errors.Init, err)
	}

	if e.visualize {
		e.graph.WriteDotString()
	}

	err = e.init()
	if err != nil {
		return err
	}

	err = e.collects()
	if err != nil {
		return err
	}

	return nil
}

// Serve used to start serving vertices
// Do not change this method fn, sync with constants in the beginning of this file
func (e *Endure) Serve() (<-chan *Result, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	e.log.Debug("preparing to serve")

	e.startMainThread()

	err := e.serve()
	if err != nil {
		return nil, err
	}

	e.log.Debug("serving")

	return e.userResultsCh, nil
}

// Stop used to shutdown the Endure
// Do not change this method fn, sync with constants in the beginning of this file
func (e *Endure) Stop() error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if len(e.graph.Vertices()) == 0 {
		return errors.E(errors.Str("no plugins registered"))
	}

	e.log.Debug("calling stop")

	return e.stop()
}

func (e *Endure) Plugins() []string {
	e.mu.RLock()
	defer e.mu.RUnlock()

	v := e.graph.TopologicalOrder()
	plugins := make([]string, 0, len(v))

	for i := range v {
		if !v[i].IsActive() {
			continue
		}

		if val, ok := v[i].Plugin().(Named); ok {
			plugins = append(plugins, val.Name())
			continue
		}

		plugins = append(plugins, v[i].ID().String())
	}

	return plugins
}

func profile() {
	go func() {
		mux := http.NewServeMux()

		mux.HandleFunc("/debug/pprof/", pprof.Index)
		mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
		mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
		mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
		mux.HandleFunc("/debug/pprof/trace", pprof.Trace)

		srv := &http.Server{
			ReadHeaderTimeout: time.Minute * 5,
			Handler:           mux,
			Addr:              "0.0.0.0:6061",
		}

		_ = srv.ListenAndServe()
	}()
}
