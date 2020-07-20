package cascade

import (
	"errors"
	"net/http"
	_ "net/http/pprof"
	"reflect"
	"sync"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/spiral/cascade/structures"
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

type Cascade struct {
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

	rwMutex *sync.RWMutex

	// result always points on healthy channel associated with vertex
	results map[string]*result

	/// main thread
	handleErrorCh chan *result
	userResultsCh chan *Result

	errorTime     map[string]*time.Time
	restartedTime map[string]*time.Time
}

type Options func(cascade *Cascade)

////////////////////////////////////////////////////////////////////////////////////////////////////////
//////////////////////////////////////////// PUBLIC ////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////////////////////////////

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
   see the cascade.Level
*/
func NewContainer(logLevel Level, options ...Options) (*Cascade, error) {
	c := &Cascade{
		rwMutex:         &sync.RWMutex{},
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
	c.errorTime = make(map[string]*time.Time)
	c.restartedTime = make(map[string]*time.Time)

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
	return func(cascade *Cascade) {
		cascade.retry = set
	}
}

func SetBackoffTimes(initialInterval time.Duration, maxInterval time.Duration) Options {
	return func(cascade *Cascade) {
		cascade.maxInterval = maxInterval
		cascade.initialInterval = initialInterval
	}
}

// Depender depends the dependencies
// name is a name of the dependency, for example - S2
// vertex is a value -> pointer to the structure
func (c *Cascade) Register(vertex interface{}) error {
	t := reflect.TypeOf(vertex)
	vertexID := removePointerAsterisk(t.String())

	if t.Kind() != reflect.Ptr {
		return errors.New("you should pass pointer to the structure instead of value")
	}

	ok := t.Implements(reflect.TypeOf((*Service)(nil)).Elem())
	if !ok {
		return typeNotImplementError
	}

	/* Depender the type
	Information we know at this step is:
	1. VertexId
	2. Vertex structure value (interface)
	And we fill vertex with this information
	*/
	err := c.register(vertexID, vertex, order)
	if err != nil {
		return err
	}
	order++
	/* Add the types, which (if) current vertex provides
	Information we know at this step is:
	1. VertexId
	2. Vertex structure value (interface)
	3. Provided type
	4. Provided type String name
	We add 3 and 4 points to the Vertex
	*/
	err = c.addProviders(vertexID, vertex)
	if err != nil {
		return err
	}
	c.logger.Debug("registering type", zap.String("type", t.String()))

	return nil
}

// Init container and all service edges.
func (c *Cascade) Init() error {
	// traverse the graph
	if err := c.addEdges(); err != nil {
		return err
	}

	// we should build init list in the reverse order
	sortedVertices := structures.TopologicalSort(c.graph.Vertices)

	if len(sortedVertices) == 0 {
		c.logger.Panic("graph should contain at least 1 vertex")
	}
	// 0 element is the HEAD
	c.runList.SetHead(&structures.DllNode{
		Vertex: sortedVertices[0]})

	// push elements to the list
	for i := 1; i < len(sortedVertices); i++ {
		c.runList.Push(sortedVertices[i])
	}

	head := c.runList.Head
	headCopy := head
	for headCopy != nil {
		err := c.init(headCopy.Vertex)
		if err != nil {
			c.logger.Error("error during the init", zap.Error(err))
			return err
		}
		headCopy = headCopy.Next
	}

	return nil
}

func (c *Cascade) Serve() (error, <-chan *Result) {
	c.rwMutex.Lock()
	defer c.rwMutex.Unlock()
	c.startMainThread()

	// call configure
	nCopy := c.runList.Head
	for nCopy != nil {
		err := c.configure(nCopy)
		if err != nil {
			c.logger.Error("backoff failed", zap.String("vertex id", nCopy.Vertex.Id), zap.Error(err))
			return err, nil
		}

		nCopy = nCopy.Next
	}

	nCopy = c.runList.Head
	for nCopy != nil {
		err := c.serve(nCopy)
		if err != nil {
			return err, nil
		}
		nCopy = nCopy.Next
	}
	return nil, c.userResultsCh
}

func (c *Cascade) Stop() error {
	c.rwMutex.Lock()
	defer c.rwMutex.Unlock()

	c.logger.Info("exiting from the Cascade")
	n := c.runList.Head
	c.shutdown(n)
	return nil
}

////////////////////////////////////////////////////////////////////////////////////////////////////////
//////////////////////////////////////////// PRIVATE ///////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////////////////////////////

func (c *Cascade) startMainThread() {
	// handle error channel goroutine
	/*
		Used for handling error from the vertices
	*/
	go func() {
		for {
			select {
			// failed Vertex
			case res, ok := <-c.handleErrorCh:
				// lock the handleErrorCh processing
				c.rwMutex.Lock()
				if !ok {
					c.logger.Debug("handle error channel was closed")
					c.rwMutex.Unlock()
					return
				}
				// received signal to exit from main goroutine
				if res.internalExit == true {
					c.rwMutex.Unlock()
					return
				}

				c.logger.Debug("processing error in the main thread", zap.String("vertex id", res.vertexId))
				if c.checkLeafErrorTime(res) {
					c.logger.Debug("error processing skipped because vertex already restartedTime by the root", zap.String("vertex id", res.vertexId))
					c.sendResultToUser(res)
					c.rwMutex.Unlock()
					continue
				}

				// get vertex from the graph
				vertex := c.graph.GetVertex(res.vertexId)
				if vertex == nil {
					c.logger.Error("failed to get vertex from the graph, vertex is nil", zap.String("vertex id from the handleErrorCh channel", res.vertexId))
					c.userResultsCh <- &Result{
						Error:    FailedToGetTheVertex,
						VertexID: "",
					}
					c.rwMutex.Unlock()
					return
				}

				// reset vertex and dependencies to the initial state
				// NumOfDeps and Visited/Visiting
				vertices := c.resetVertices(vertex)

				// Topologically sort the graph
				sorted := structures.TopologicalSort(vertices)
				if sorted == nil {
					c.logger.Error("sorted list should not be nil", zap.String("vertex id from the handleErrorCh channel", res.vertexId))
					c.userResultsCh <- &Result{
						Error:    FailedToSortTheGraph,
						VertexID: "",
					}
					c.rwMutex.Unlock()
					return
				}

				if c.retry {
					// send exit signal only to sorted and involved vertices
					c.sendExitSignal(sorted)

					// Init backoff
					b := backoff.NewExponentialBackOff()
					b.MaxInterval = c.maxInterval
					b.InitialInterval = c.initialInterval

					affectedRunList := structures.NewDoublyLinkedList()
					affectedRunList.SetHead(&structures.DllNode{
						Vertex: sorted[len(sorted)-1]})

					for i := len(sorted) - 2; i >= 0; i-- {
						affectedRunList.Push(sorted[i])
					}

					// call init
					headCopy := affectedRunList.Head
					for headCopy != nil {
						berr := backoff.Retry(c.backoffInit(headCopy.Vertex), b)
						if berr != nil {
							c.logger.Error("backoff failed", zap.String("vertex id", headCopy.Vertex.Id), zap.Error(berr))
							c.userResultsCh <- &Result{
								Error:    ErrorDuringInit,
								VertexID: headCopy.Vertex.Id,
							}
							c.rwMutex.Unlock()
							return
						}

						headCopy = headCopy.Next
					}

					// call configure
					headCopy = affectedRunList.Head
					for headCopy != nil {
						berr := backoff.Retry(c.backoffConfigure(headCopy), b)
						if berr != nil {
							c.userResultsCh <- &Result{
								Error:    ErrorDuringInit,
								VertexID: headCopy.Vertex.Id,
							}
							c.logger.Error("backoff failed", zap.String("vertex id", headCopy.Vertex.Id), zap.Error(berr))
							c.rwMutex.Unlock()
							return
						}

						headCopy = headCopy.Next
					}

					// call serve
					headCopy = affectedRunList.Head
					for headCopy != nil {
						err := c.serve(headCopy)
						if err != nil {
							c.userResultsCh <- &Result{
								Error:    ErrorDuringServe,
								VertexID: headCopy.Vertex.Id,
							}
							c.logger.Error("fatal error during the serve in the main thread", zap.String("vertex id", headCopy.Vertex.Id), zap.Error(err))
							c.rwMutex.Unlock()
							return
						}

						headCopy = headCopy.Next
					}

					c.sendResultToUser(res)
					c.rwMutex.Unlock()
				} else {
					c.logger.Info("retry is turned off, sending exit signal to every vertex in the graph")
					// send exit signal to whole graph
					c.sendExitSignal(c.graph.Vertices)
					c.sendResultToUser(res)
					c.rwMutex.Unlock()
				}
			}
		}
	}()
}
