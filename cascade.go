package cascade

import (
	"context"
	"os"
	"reflect"

	"github.com/rs/zerolog"
	"github.com/spiral/cascade/structures"
)

type Cascade struct {
	// Dependency graph
	graph *structures.Graph
	// DLL used as run list to run in order
	runList *structures.DoublyLinkedList
	// logger
	logger zerolog.Logger

	ctx    context.Context
	cancel context.CancelFunc
}

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

////////////////////////////////////////////////////////////////////////////////////////////////////////
//////////////////////////////////////////// PUBLIC ////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////////////////////////////

/* Input parameters: logLevel
-1 is the most informative level - TraceLevel
0 - DebugLevel defines debug log level
1 - InfoLevel defines info log level.
2 - WarnLevel defines warn log level.
3 - ErrorLevel defines error log level.
4 - FatalLevel defines fatal log level.
5 - PanicLevel defines panic log level.
6 - NoLevel defines an absent log level.
7 - Disabled disables the logger.
see the cascade.Level
*/
func NewContainer(logLevel Level) (*Cascade, error) {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix

	switch logLevel {
	case DebugLevel:
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
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
	ctx, stop := context.WithCancel(context.Background())

	return &Cascade{
		graph:   structures.NewGraph(),
		runList: structures.NewDoublyLinkedList(),
		logger:  logger,
		cancel:  stop,
		ctx:     ctx,
	}, nil
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

	c.logger.Info().Msgf("registered type: %s", t.String())

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
	sortedVertices := c.graph.TopologicalSort()

	// TODO properly handle the len of the sorted vertices
	c.runList.SetHead(&structures.DllNode{
		Vertex: sortedVertices[0]})

	// TODO what if sortedVertices will contain only 1 node (len(sortedVertices) - 2 will panic)
	for i := 1; i < len(sortedVertices); i++ {
		c.runList.Push(sortedVertices[i])
	}

	err := c.init(c.runList.Head)
	if err != nil {
		c.logger.
			Err(err).
			Stack().
			Msg("error during the init")
	}
	return c.init(c.runList.Head)
}

func (c *Cascade) Configure() error {
	return nil
}

func (c *Cascade) Close() error {
	return nil
}

func (c *Cascade) Serve() <-chan *Result {
	//ch := make(chan error, 1)
	n := c.runList.Head
	//go func() {
	//
	//	ch
	//}()
	//r := Result{
	//	ErrCh:    make(chan error, len(c.graph.Vertices)),
	//	VertexID: "",
	//}
	//go func() {
	//	for k := range  {
	//		if k != nil {
	//
	//		}
	//	}
	//}()

	// restart vertices on fail



	return merge(c.internalServe(n))
}

func (c *Cascade) Stop() error {
	n := c.runList.Head

	for n != nil {
		err := c.internalStop(n)
		if err != nil {
			// TODO do not return until finished
			// just log the errors
			// stack it in slice and if slice is not empty, print it ??
			c.logger.Err(err).Stack().Msg("error occurred during the services stopping")
		}

		// prev DLL node
		n = n.Next
	}
	return nil
}

func (c *Cascade) Get(name string) interface{} {
	panic("unimplemented!")
}
func (c *Cascade) Has(name string) bool {
	panic("unimplemented!")
}

func (c *Cascade) List() []string {
	panic("unimplemented!")
	return nil
}

////////////////////////////////////////////////////////////////////////////////////////////////////////
//////////////////////////////////////////// PRIVATE ///////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////////////////////////////

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
func (c *Cascade) init(n *structures.DllNode) error {
	// traverse the dll
	for n != nil {
		// we already checked the Interface satisfaction
		// at this step absence of Init() is impossible
		init, _ := reflect.TypeOf(n.Vertex.Iface).MethodByName(InitMethodName)

		initArgs, err := functionParameters(init)
		if err != nil {
			return err
		}

		// If len(initArgs) is eq to 1, than we deal with empty Init() method
		//
		if len(initArgs) == 1 {
			err = c.noDepsCall(init, n)
			if err != nil {
				c.logger.
					Err(err).
					Stack().Str("vertexID", n.Vertex.Id).
					Msg("error in noDepsCall")
				return err
			}
		} else {
			// else, we deal with variadic len of Init function parameters Init(a,b,c, etc)
			// we should resolve all it all
			err = c.depsCall(init, n)
			if err != nil {
				c.logger.
					Err(err).
					Stack().Str("vertexID", n.Vertex.Id).
					Msg("error in depsCall")
				return err
			}
		}

		// next DLL node
		n = n.Next
	}

	return nil
}

// stopReverse will call Stop on every node in node.Prev in DLL
func (c *Cascade) stopReverse(n *structures.DllNode) error {
	// traverse the dll
	for n != nil {
		in := make([]reflect.Value, 0, 1)
		// add service itself
		in = append(in, reflect.ValueOf(n.Vertex.Iface))

		err := c.close(n, in)
		if err != nil {
			c.logger.Err(err).Stack().Msg("error occurred during the services closing")
		}
		err = c.internalStop(n)
		if err != nil {
			// TODO do not return until finished
			// just log the errors
			// stack it in slice and if slice is not empty, print it ??
			c.logger.Err(err).Stack().Msg("error occurred during the services stopping")
		}

		// prev DLL node
		n = n.Prev
	}

	return nil
}
