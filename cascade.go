package cascade

import (
	"log"
	"net/http"
	_ "net/http/pprof"
	"reflect"
	"sync"
	"time"

	"github.com/spiral/cascade/structures"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
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

type Cascade struct {
	// Dependency graph
	graph *structures.Graph
	// DLL used as run list to run in order
	runList *structures.DoublyLinkedList
	// logger
	logger *zap.Logger
	// OPTIONS
	// retry on vertex fail
	retry bool
	// retry if init fails, by default this fatal issue
	// when INIT fails, there is smt wrong with application
	retryOnInitFail bool
	// retry in case of configure issues
	retryOnConfigureFail bool
	numOfRetries         int

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
   -1 is the most informative level - TraceLevel --> also turns on pprof endpoint
   0 - DebugLevel defines debug log level --> also turns on pprof endpoint
   1 - InfoLevel defines info log level.
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
		rwMutex: &sync.RWMutex{},
	}

	var lvl zap.AtomicLevel
	switch logLevel {
	case DebugLevel:
		lvl = zap.NewAtomicLevelAt(zap.DebugLevel)
		// start pprof
		startPprof()
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
			MessageKey: "message",

			LevelKey:    "level",
			EncodeLevel: zapcore.CapitalLevelEncoder,

			TimeKey:    "time",
			EncodeTime: zapcore.ISO8601TimeEncoder,

			CallerKey:    "caller",
			EncodeCaller: zapcore.ShortCallerEncoder,
		},
		OutputPaths:      []string{"stderr"},
		ErrorOutputPaths: []string{"stderr"},
	}

	logger, err := cfg.Build()
	if err != nil {
		return nil, err
	}
	c.logger = logger

	// append options
	for _, option := range options {
		option(c)
	}

	c.graph = structures.NewGraph()
	c.runList = structures.NewDoublyLinkedList()
	c.logger = logger
	c.results = make(map[string]*result)

	// FINAL
	c.handleErrorCh = make(chan *result)
	c.userResultsCh = make(chan *Result)
	c.errorTime = make(map[string]*time.Time)
	c.restartedTime = make(map[string]*time.Time)

	return c, nil
}

func startPprof() {
	go func() {
		log.Println(http.ListenAndServe("0.0.0.0:6061", nil))
	}()
}

func RetryOnFail(set bool) Options {
	return func(cascade *Cascade) {
		cascade.retry = set
		// default value
		cascade.numOfRetries = 5
	}
}

// Register depends the dependencies
// name is a name of the dependency, for example - S2
// vertex is a value -> pointer to the structure
func (c *Cascade) Register(vertex interface{}) error {
	t := reflect.TypeOf(vertex)
	vertexID := removePointerAsterisk(t.String())

	ok := t.Implements(reflect.TypeOf((*Service)(nil)).Elem())
	if !ok {
		return typeNotImplementError
	}

	/* Register the type
	Information we know at this step is:
	1. VertexId
	2. Vertex structure value (interface)
	And we fill vertex with this information
	*/
	err := c.register(vertexID, vertex)
	if err != nil {
		return err
	}

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
	sortedVertices := structures.OldTopologicalSort(c.graph.Vertices)

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
	n := c.runList.Head
	err := c.serveRunList(n)
	if err != nil {
		return err, nil
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

func (c *Cascade) restart() error {
	c.handleErrorCh <- &result{
		internalExit: true,
	}

	c.rwMutex.Lock()
	defer c.rwMutex.Unlock()

	c.logger.Info("restarting the Cascade")
	n := c.runList.Head

	// shutdown, send exit signals to every user Serve() goroutine
	c.shutdown(n)

	// reset the run list to initial state
	c.runList.Reset()
	// reset all results
	c.results = make(map[string]*result)
	// reset error timings
	c.errorTime = make(map[string]*time.Time)
	// reset restarted timings
	c.restartedTime = make(map[string]*time.Time)

	// re-start main thread
	c.startMainThread()

	// error only can be received from the configure method
	err := c.serveRunList(n)
	if err != nil {
		return err
	}
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

				for _, v := range sorted {
					// skip self
					if v.Id == res.vertexId {
						continue
					}
					// get result by vertex ID
					tmp := c.results[v.Id]
					// send exit signal to the goroutine in sorted order
					c.logger.Debug("sending exit signal to the vertex in the main thread", zap.String("vertex id", tmp.vertexId))
					tmp.exit <- struct{}{}

					c.results[v.Id] = nil
				}

				c.logger.Debug("sending exit signal to the vertex in the main thread", zap.String("vertex id", res.vertexId))
				// close self here
				res.exit <- struct{}{}

				if c.retry {
					// TODO --> to the separate function
					// creating run list
					affectedRunList := structures.NewDoublyLinkedList()
					// TODO properly handle the len of the sorted vertices
					affectedRunList.SetHead(&structures.DllNode{
						Vertex: sorted[len(sorted)-1]})

					// TODO what if sortedVertices will contain only 1 node (len(sortedVertices) - 2 will panic)
					for i := len(sorted) - 2; i >= 0; i-- {
						affectedRunList.Push(sorted[i])
					}

					head := affectedRunList.Head
					headCopy := head

					for headCopy != nil {
						err := c.init(headCopy.Vertex)
						if err != nil {
							c.logger.Error("error while invoke Init", zap.String("vertex id", head.Vertex.Id), zap.Error(err))
							c.userResultsCh <- &Result{
								Error:    ErrorDuringInit,
								VertexID: headCopy.Vertex.Id,
							}
							c.rwMutex.Unlock()
							return
						}

						headCopy = headCopy.Next
					}

					// start serving run list, error can be returned only from the configure
					err := c.serveRunList(head)
					if err != nil {
						c.userResultsCh <- &Result{
							Error:    ErrorDuringInit,
							VertexID: headCopy.Vertex.Id,
						}
						c.logger.Error("fatal error during the configure in main thread", zap.String("vertex id", head.Vertex.Id), zap.Error(err))
						c.rwMutex.Unlock()
						return
					}
				}
				c.sendResultToUser(res)
				c.rwMutex.Unlock()
			}
		}
	}()
}

func (c *Cascade) sendResultToUser(res *result) {
	c.userResultsCh <- &Result{
		Error: Error{
			Err: res.err,
		},
		VertexID: res.vertexId,
	}
}

func (c *Cascade) shutdown(n *structures.DllNode) {
	for n != nil {
		err := c.internalStop(n.Vertex.Id)
		if err != nil {
			// TODO do not return until finished
			// just log the errors
			// stack it in slice and if slice is not empty, print it ??
			c.logger.Error("error occurred during the services stopping", zap.String("vertex id", n.Vertex.Id), zap.Error(err))
		}
		if channel, ok := c.results[n.Vertex.Id]; ok && channel != nil {
			channel.exit <- struct{}{}
		}

		// next DLL node
		n = n.Next
	}
}

// serveRunList run configure (if exist) and serve for each node and put the results in the map
func (c *Cascade) serveRunList(n *structures.DllNode) error {
	nCopy := n

	for nCopy != nil {
		// handle all configure
		in := make([]reflect.Value, 0, 1)
		// add service itself
		in = append(in, reflect.ValueOf(nCopy.Vertex.Iface))

		//var res Result
		if reflect.TypeOf(nCopy.Vertex.Iface).Implements(reflect.TypeOf((*Graceful)(nil)).Elem()) {
			// TODO backoff here
			err := c.configure(nCopy.Vertex, in)
			if err != nil {
				return err
			}
		}

		res := c.serve(nCopy.Vertex, in)
		if res != nil {
			c.results[res.vertexId] = res
		} else {
			c.logger.Panic("nil result returned from the vertex", zap.String("vertex id", nCopy.Vertex.Id))
		}

		c.poll(res)
		if c.restartedTime[nCopy.Vertex.Id] != nil {
			*c.restartedTime[nCopy.Vertex.Id] = time.Now()
		} else {
			tmp := time.Now()
			c.restartedTime[nCopy.Vertex.Id] = &tmp
		}

		nCopy = nCopy.Next
	}

	return nil
}

func (c *Cascade) checkLeafErrorTime(res *result) bool {
	return c.restartedTime[res.vertexId] != nil && (*c.restartedTime[res.vertexId]).After(*c.errorTime[res.vertexId])
}

// poll is used to poll the errors from the vertex
// and exit from it
func (c *Cascade) poll(r *result) {
	rr := r
	go func(res *result) {
		for {
			select {
			// error
			case e := <-res.errCh:
				if e != nil {
					// set error time
					c.rwMutex.Lock()
					if c.errorTime[res.vertexId] != nil {
						*c.errorTime[res.vertexId] = time.Now()
					} else {
						tmp := time.Now()
						c.errorTime[res.vertexId] = &tmp
					}
					c.rwMutex.Unlock()

					c.logger.Error("error processed in poll", zap.String("vertex id", res.vertexId), zap.Error(e))

					// set the error
					res.err = e

					// send handleErrorCh signal
					c.handleErrorCh <- res
				}
			// exit from the goroutine
			case <-res.exit:
				c.logger.Info("got exit signal", zap.String("vertex id", res.vertexId))
				err := c.internalStop(res.vertexId)
				if err != nil {
					c.logger.Error("error during exit signal", zap.String("error while stopping the vertex:", res.vertexId), zap.Error(err))
				}
				return
			}
		}
	}(rr)
}

// TODO graph responsibility, not Cascade
func (c *Cascade) resetVertices(vertex *structures.Vertex) []*structures.Vertex {
	// restore number of dependencies for the root
	vertex.NumOfDeps = len(vertex.Dependencies)
	vertex.Visiting = false
	vertex.Visited = false
	vertices := make([]*structures.Vertex, 0, 5)
	vertices = append(vertices, vertex)

	tmp := make(map[string]*structures.Vertex)

	c.dfs(vertex.Dependencies, tmp)

	for _, v := range tmp {
		vertices = append(vertices, v)
	}
	return vertices
}

// TODO graph responsibility, not Cascade
func (c *Cascade) dfs(deps []*structures.Vertex, tmp map[string]*structures.Vertex) {
	for i := 0; i < len(deps); i++ {
		deps[i].Visited = false
		deps[i].Visiting = false
		deps[i].NumOfDeps = len(deps)
		tmp[deps[i].Id] = deps[i]
		c.dfs(deps[i].Dependencies, tmp)
	}
}

func (c *Cascade) register(name string, vertex interface{}) error {
	// check the vertex
	if c.graph.HasVertex(name) {
		return vertexAlreadyExists(name)
	}

	// just push the vertex
	// here we can append in future some meta information
	c.graph.AddVertex(name, vertex, structures.Meta{})
	return nil
}

/*
   Traverse the DLL in the forward direction

*/
func (c *Cascade) init(v *structures.Vertex) error {
	// we already checked the Interface satisfaction
	// at this step absence of Init() is impossible
	init, _ := reflect.TypeOf(v.Iface).MethodByName(InitMethodName)

	err := c.initCall(init, v)
	if err != nil {
		c.logger.Error("error occurred while calling a function", zap.String("vertex id", v.Id), zap.Error(err))
		return err
	}

	return nil
}
