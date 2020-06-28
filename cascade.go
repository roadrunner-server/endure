package cascade

import (
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"reflect"
	"sync"
	"time"

	"github.com/rs/zerolog"
	"github.com/spiral/cascade/structures"
)

// Level defines log levels.
type Level int8

const (
	// DebugLevel defines debug log level.
	DebugLevel Level = iota
	// InfoLevel defines info log level.
	InfoLevel
	// WarnLevel defines warn log level.
	WarnLevel
	// ErrorLevel defines error log level.
	ErrorLevel
	// FatalLevel defines fatal log level.
	FatalLevel
	// PanicLevel defines panic log level.
	PanicLevel
	// NoLevel defines an absent log level.
	NoLevel
	// Disabled disables the logger.
	Disabled

	// TraceLevel defines trace log level.
	TraceLevel Level = -1
)

type Cascade struct {
	// Dependency graph
	graph *structures.Graph
	// DLL used as run list to run in order
	runList *structures.DoublyLinkedList
	// logger
	logger zerolog.Logger
	// OPTIONS
	retryOnFail bool
	//retryFunc    func(serveChannels map[string]*result, restart chan struct{}) chan *Result
	numOfRetries int

	rwMutex *sync.RWMutex

	// result always points on healthy channel associated with vertex
	results map[string]*result

	/// main thread
	handleErrorCh chan *result
	userResultsCh chan *Result
	shutdownCh    chan struct{}
	errorTimings  map[string]*time.Time
	restarted     map[string]*time.Time
}

type Options func(cascade *Cascade)

////////////////////////////////////////////////////////////////////////////////////////////////////////
//////////////////////////////////////////// PUBLIC ////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////////////////////////////

/* Input parameters: logLevel
-1 is the most informative level - TraceLevel
0 - DebugLevel defines debug log level --> also turn on pprof endpoint
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
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix

	switch logLevel {
	case DebugLevel:
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
		// start pprof
		startPprof()
	case InfoLevel:
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	case WarnLevel:
		zerolog.SetGlobalLevel(zerolog.WarnLevel)
	case ErrorLevel:
		zerolog.SetGlobalLevel(zerolog.ErrorLevel)
	case FatalLevel:
		zerolog.SetGlobalLevel(zerolog.FatalLevel)
	case PanicLevel:
		zerolog.SetGlobalLevel(zerolog.PanicLevel)
	case NoLevel:
		zerolog.SetGlobalLevel(zerolog.NoLevel)
	case Disabled:
		zerolog.SetGlobalLevel(zerolog.Disabled)
	case -1: // TraceLevel
		zerolog.SetGlobalLevel(zerolog.TraceLevel)
	default:
		zerolog.SetGlobalLevel(zerolog.Disabled)
	}

	logger := zerolog.New(os.Stderr).With().Timestamp().Logger()

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
	c.shutdownCh = make(chan struct{})
	c.errorTimings = make(map[string]*time.Time)
	c.restarted = make(map[string]*time.Time)

	return c, nil
}

func startPprof() {
	go func() {
		log.Println(http.ListenAndServe("0.0.0.0:6061", nil))
	}()
}

func RetryOnFail(set bool) Options {
	return func(cascade *Cascade) {
		cascade.retryOnFail = set
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

	c.logger.Info().Str("type", t.String()).Msgf("registering type")

	return nil
}

// Init container and all service edges.
func (c *Cascade) Init() error {
	// traverse the graph
	if err := c.addEdges(); err != nil {
		return err
	}

	// we should buld init list in the reverse order
	// TODO return cycle error
	sortedVertices := structures.OldTopologicalSort(c.graph.Vertices)

	// TODO properly handle the len of the sorted vertices
	c.runList.SetHead(&structures.DllNode{
		Vertex: sortedVertices[0]})

	// TODO what if sortedVertices will contain only 1 node (len(sortedVertices) - 2 will panic)
	for i := 1; i < len(sortedVertices); i++ {
		c.runList.Push(sortedVertices[i])
	}

	head := c.runList.Head
	headCopy := head
	for headCopy != nil {
		err := c.init(headCopy.Vertex)
		if err != nil {
			c.logger.
				Err(err).
				Stack().
				Msg("error during the init")
			return err
		}
		headCopy = headCopy.Next
	}

	return nil
}

func (c *Cascade) Configure() error {
	return nil
}

func (c *Cascade) Close() error {
	return nil
}

func (c *Cascade) Serve() <-chan *Result {
	c.startMainThread()
	n := c.runList.Head
	c.rwMutex.Lock()
	for n != nil {
		// initial start

		res := c.serveVertex(n.Vertex)
		// if err != nil, but we set up restart
		if res != nil {
			c.results[res.vertexId] = res
		} else {
			panic("res nil")
		}

		c.poll(res)
		if c.restarted[n.Vertex.Id] != nil {
			*c.restarted[n.Vertex.Id] = time.Now()
		} else {
			tmp := time.Now()
			c.restarted[n.Vertex.Id] = &tmp
		}

		n = n.Next
	}
	c.rwMutex.Unlock()

	//next listen for the failing nodes after start
	if c.retryOnFail {
		return c.userResultsCh
	}

	return c.userResultsCh
}

func (c *Cascade) Stop() error {

	c.shutdownCh <- struct {}{}

	return nil
}

////////////////////////////////////////////////////////////////////////////////////////////////////////
//////////////////////////////////////////// PRIVATE ///////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////////////////////////////

func (c *Cascade) startMainThread() {
	go func() {
		for {
			select {
			// failed Vertex
			case res, ok := <-c.handleErrorCh:
				if !ok {
					c.logger.Info().Msg("handle error channel was closed")
					return
				}
				c.logger.Info().Str("vertex id", res.vertexId).Msg("processing error in the main thread")
				if c.checkLeafErrorTime(res) {
					c.logger.Info().Str("vertex id", res.vertexId).Msg("error processing skipped because vertex already restarted by the root")
					break
				}

				// lock the handleErrorCh processing
				c.rwMutex.Lock()

				// get vertex from the graph
				vertex := c.graph.GetVertex(res.vertexId)
				if vertex == nil {
					c.logger.Fatal().Str("vertex id from the handleErrorCh channel", res.vertexId).Msg("failed to get vertex from the graph, vertex is nil")
				}

				// reset vertex and dependencies to the initial state
				// NumOfDeps and Visited/Visiting
				vertices := c.resetVertices(vertex)

				// Topologically sort the graph
				sorted := structures.TopologicalSort(vertices)
				if sorted == nil {
					c.logger.Fatal().Str("vertex id from the handleErrorCh channel", res.vertexId).Msg("sorted list should not be nil")
				}

				for _, v := range sorted {
					// skip self
					if v.Id == res.vertexId {
						continue
					}
					// get result by vertex ID
					tmp := c.results[v.Id]
					// send exit signal to the goroutine in sorted order
					c.logger.Info().Str("vertex id", tmp.vertexId).Msg("sending exit signal to the vertex in the main thread")
					tmp.exit <- struct{}{}

					c.results[v.Id] = nil
				}

				c.logger.Info().Str("vertex id", res.vertexId).Msg("sending exit signal to the vertex in the main thread")
				// close self here
				res.exit <- struct{}{}

				if c.retryOnFail {
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
							c.logger.
								Fatal().
								Err(err).
								Stack().
								Msg("error during the startMainThread")
						}
						headCopy = headCopy.Next
					}

					for head != nil {
						// serve current vertex
						// TODO backoff
						r := c.serveVertex(head.Vertex)
						// if err != nil, but we set up restart
						if r != nil {
							c.results[r.vertexId] = r
						} else {
							panic("res nil")
						}

						// start polling events from the vertex
						c.poll(r)
						// set restarted time
						if c.restarted[head.Vertex.Id] != nil {
							*c.restarted[head.Vertex.Id] = time.Now()
						} else {
							tmp := time.Now()
							c.restarted[head.Vertex.Id] = &tmp
						}

						// move to the next node
						head = head.Next
					}
				}

				// unlock the scope
				c.rwMutex.Unlock()
			case <-c.shutdownCh:
				c.logger.Info().Msg("exiting from the Cascade")
				n := c.runList.Head

				c.rwMutex.Lock()
				for n != nil {
					err := c.internalStop(n.Vertex.Id)
					if err != nil {
						// TODO do not return until finished
						// just log the errors
						// stack it in slice and if slice is not empty, print it ??
						c.logger.Err(err).Stack().Msg("error occurred during the services stopping")
					}
					if channel, ok := c.results[n.Vertex.Id]; ok && channel != nil {
						channel.exit <- struct{}{}
					}

					// prev DLL node
					n = n.Next

				}
				c.rwMutex.Unlock()
				//close(c.handleErrorCh)
			}
		}
	}()
}

func (c *Cascade) checkLeafErrorTime(res *result) bool {
	c.rwMutex.Lock()
	defer c.rwMutex.Unlock()
	return c.restarted[res.vertexId] != nil && (*c.restarted[res.vertexId]).After(*c.errorTimings[res.vertexId])
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
					if c.errorTimings[res.vertexId] != nil {
						*c.errorTimings[res.vertexId] = time.Now()
					} else {
						tmp := time.Now()
						c.errorTimings[res.vertexId] = &tmp
					}
					c.rwMutex.Unlock()

					c.logger.Err(e).
						Str("error occurred in the vertex:", res.vertexId).
						Msg("error processed in poll")

					c.userResultsCh <- &Result{
						Err:      e,
						Code:     0,
						VertexID: res.vertexId,
					}

					// send handleErrorCh signal
					c.handleErrorCh <- res
				}
				// exit from the goroutine
			case <-res.exit:
				c.logger.Log().
					Str("exiting:", res.vertexId).
					Msg("got exit signal")
				err := c.internalStop(res.vertexId)
				if err != nil {
					c.logger.Err(err).
						Str("error while stopping the vertex:", res.vertexId).
						Msg("error during exit signal")
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
		c.logger.
			Err(err).
			Stack().Str("vertexID", v.Id).
			Msg("error occurred while calling a function")
		return err
	}

	return nil
}
