package cascade

import (
	"errors"
	"fmt"
	"os"
	"reflect"
	"strings"

	"github.com/rs/zerolog"
	"github.com/spiral/cascade/structures"
)

// InitMethodName is the function name for the reflection
const InitMethodName = "Init"
// Stop is the function name for the reflection to Stop the service
const StopMethodName = "Stop"

type Cascade struct {
	// Dependency graph
	graph *structures.Graph
	// DLL used as run list to run in order
	runList *structures.DoublyLinkedList
	// logger
	logger zerolog.Logger
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

	return &Cascade{
		graph:   structures.NewGraph(),
		runList: structures.NewDoublyLinkedList(),
		logger:  logger,
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
	if err := c.calculateEdges(); err != nil {
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

	return c.init(c.runList.Head)
}

func (c *Cascade) Serve(upstream chan interface{}) error {
	panic("unimplemented!")
	return nil
}
func (c *Cascade) Stop() error {
	panic("unimplemented!")
	return nil
}

func (c *Cascade) Get(name string) interface{} {
	panic("unimplemented!")
	return nil
}
func (c *Cascade) Has(name string) bool {
	panic("unimplemented!")
	return false
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

func (c *Cascade) addProviders(vertexID string, vertex interface{}) error {
	if provider, ok := vertex.(Provider); ok {
		for _, fn := range provider.Provides() {
			ret, err := providersReturnType(fn)
			if err != nil {
				// todo: delete gVertex
				return err
			}

			typeStr := removePointerAsterisk(ret.String())
			// get the Vertex from the graph (gVertex)
			gVertex := c.graph.GetVertex(vertexID)
			if gVertex.Provides == nil {
				gVertex.Provides = make(map[string]structures.ProvidedEntry)
			}

			if gVertex.Meta.FnsProviderToInvoke == nil {
				gVertex.Meta.FnsProviderToInvoke = make([]string, 0, 5)
			}

			gVertex.Meta.FnsProviderToInvoke = append(gVertex.Meta.FnsProviderToInvoke, functionName(fn))

			gVertex.Provides[typeStr] = structures.ProvidedEntry{
				IsReference: nil,
				Value:       nil,
			}
		}
	}
	return nil
}

// calculateEdges calculates simple graph for the dependencies
func (c *Cascade) calculateEdges() error {
	// vertexID for example S2
	for vertexID, vrtx := range c.graph.Graph {
		init, ok := reflect.TypeOf(vrtx.Iface).MethodByName(InitMethodName)
		if !ok {
			panic("init method should be implemented")
		}

		/* Add the dependencies (if) which this vertex needs to init
		Information we know at this step is:
		1. VertexId
		2. Vertex structure value (interface)
		3. Provided type
		4. Provided type String name
		5. Name of the dependencies which we should found
		We add 3 and 4 points to the Vertex
		*/
		err := c.calculateRegisterDeps(vertexID, vrtx.Iface)
		if err != nil {
			return err
		}

		/*
			At this step we know (and build) all dependencies via the Depends interface and connected all providers
			to it's dependencies.
			The next step is to calculate dependencies provided by the Init() method
			for example S1.Init(foo2.DB) S1 --> foo2.S2 (not foo2.DB, because vertex which provides foo2.DB is foo2.S2)
		*/
		err = c.calculateInitDeps(vertexID, init)
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *Cascade) calculateRegisterDeps(vertexID string, vertex interface{}) error {
	if register, ok := vertex.(Register); ok {
		for _, fn := range register.Depends() {
			// what type it might depend on?
			argsTypes, err := argType(fn)
			if err != nil {
				// todo: delete vertex
				return err
			}

			// empty Depends, show warning and continue
			// should be at least to
			// TODO correct
			if len(argsTypes) == 0 {
				fmt.Printf("%s must accept exactly one argument", fn)
			}

			// TODO correct
			if len(argsTypes) > 0 {
				// at is like foo2.S2
				for _, at := range argsTypes {
					// check if type is primitive type
					// TODO show warning, because why to receive primitive type in Init() ??? Any sense?
					if isPrimitive(at.String()) {
						continue
					}
					atStr := at.String()
					if vertexID == atStr {
						continue
					}
					// if we found, that some structure depends on some type
					// we also save it in the `depends` section
					// name s1 (for example)
					// vertex - S4 func

					// we store pointer in the Deps structure in the isRef field
					c.graph.AddDep(vertexID, removePointerAsterisk(atStr), structures.Depends, isReference(at))
					c.logger.Info().
						Str("vertexID", vertexID).
						Str("depends", atStr).
						Msg("adding dependency via Depends()")
				}
			} else {
				// todo temporary
				panic("argsTypes less than 0")
			}

			// get the Vertex from the graph (gVertex)
			gVertex := c.graph.GetVertex(vertexID)
			if gVertex.Provides == nil {
				gVertex.Provides = make(map[string]structures.ProvidedEntry)
			}

			if gVertex.Meta.FnsRegisterToInvoke == nil {
				gVertex.Meta.FnsRegisterToInvoke = make([]string, 0, 5)
			}

			gVertex.Meta.FnsRegisterToInvoke = append(gVertex.Meta.FnsRegisterToInvoke, functionName(fn))
		}
	}

	return nil
}

func (c *Cascade) calculateInitDeps(vertexID string, initMethod reflect.Method) error {
	// S2 init args
	initArgs, err := functionParameters(initMethod)
	if err != nil {
		return err
	}

	// iterate over all function parameters
	for _, initArg := range initArgs {
		// receiver
		if vertexID == removePointerAsterisk(initArg.String()) {
			continue
		}

		c.graph.AddDep(vertexID, removePointerAsterisk(initArg.String()), structures.Init, isReference(initArg))
		c.logger.Info().
			Str("vertexID", vertexID).
			Str("depends", initArg.String()).
			Msg("adding dependency via Init()")
	}
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
				err2 := c.runBackward(n.Prev)
				if err2 != nil {
					panic(err2)
				}
				return err
			}
		} else {
			// else, we deal with variadic len of Init function parameters Init(a,b,c, etc)
			// we should resolve all it all
			err = c.depsCall(init, n)
			if err != nil {
				err2 := c.runBackward(n.Prev)
				if err2 != nil {
					panic(err2)
				}
				return err
			}
		}

		// next DLL node
		n = n.Next
	}

	return nil
}

func (c *Cascade) runBackward(n *structures.DllNode) error {
	c.logger.Info().Msg("running backward")
	// traverse the dll
	for n != nil {
		// we already checked the Interface satisfaction
		// at this step absence of Stop() is impossible
		stop, _ := reflect.TypeOf(n.Vertex.Iface).MethodByName(StopMethodName)

		in := make([]reflect.Value, 0, 1)

		// add service itself, this is only 1 dependency for the Stop
		in = append(in, reflect.ValueOf(n.Vertex.Iface))

		ret := stop.Func.Call(in)
		rErr := ret[0].Interface()
		if rErr != nil {
			e := rErr.(error)
			panic(e)
		}

		// prev DLL node
		n = n.Prev
	}

	return nil
}

func (c *Cascade) noDepsCall(init reflect.Method, n *structures.DllNode) error {
	in := make([]reflect.Value, 0, 1)

	// add service itself
	in = append(in, reflect.ValueOf(n.Vertex.Iface))

	ret := init.Func.Call(in)
	rErr := ret[0].Interface()
	if rErr != nil {
		if e, ok := rErr.(error); ok {
			c.logger.Err(e)
			return e
		} else {
			return errors.New("unknown error occurred during the function call")
		}
	}
	// just to be safe here
	if len(in) > 0 {
		// `in` type here is initialized function receiver
		c.logger.Info().
			Str("vertexID", n.Vertex.Id).
			Str("in parameter", in[0].Type().String()).
			Msg("calling with no deps")
		err := n.Vertex.AddValue(removePointerAsterisk(in[0].Type().String()), in[0], isReference(in[0].Type()))
		if err != nil {
			return err
		}
	}

	err := c.traverseProvider(n, in)
	if err != nil {
		c.logger.Err(err)
		return err
	}

	err = c.traverseRegisters(n)
	if err != nil {
		c.logger.Err(err)
		return err
	}

	return nil
}

func (c *Cascade) traverseRegisters(n *structures.DllNode) error {
	inReg := make([]reflect.Value, 0, 1)

	// add service itself
	inReg = append(inReg, reflect.ValueOf(n.Vertex.Iface))

	// add dependencies
	if len(n.Vertex.Meta.DepsList) > 0 {
		for i := 0; i < len(n.Vertex.Meta.DepsList); i++ {
			depId := n.Vertex.Meta.DepsList[i].Name
			v := c.graph.FindProvider(depId)

			for k, val := range v.Provides {
				if k == depId {
					// value - reference and init dep also reference
					if *val.IsReference == *n.Vertex.Meta.DepsList[i].IsReference {
						inReg = append(inReg, *val.Value)
					} else if *val.IsReference {
						// same type, but difference in the refs
						// Init needs to be a value
						// But Vertex provided reference

						inReg = append(inReg, val.Value.Elem())
					} else if !*val.IsReference {
						// vice versa
						// Vertex provided value
						// but Init needs to be a reference
						if val.Value.CanAddr() {
							inReg = append(inReg, val.Value.Addr())
						} else {
							c.logger.Warn().Str("type", val.Value.Type().String()).Msgf("value is not addressible. TIP: consider to return a pointer from %s", val.Value.Type())
							c.logger.Warn().Msgf("making a fresh pointer")

							nt := reflect.New(val.Value.Type())
							inReg = append(inReg, nt)
						}
					}
				}
			}
		}
	}

	//type implements Register interface
	if reflect.TypeOf(n.Vertex.Iface).Implements(reflect.TypeOf((*Register)(nil)).Elem()) {
		// if type implements Register() it should has FnsProviderToInvoke
		if n.Vertex.Meta.DepsList != nil {
			for i := 0; i < len(n.Vertex.Meta.FnsRegisterToInvoke); i++ {
				m, ok := reflect.TypeOf(n.Vertex.Iface).MethodByName(n.Vertex.Meta.FnsRegisterToInvoke[i])
				if !ok {
					panic("method Provides should be")
				}

				ret := m.Func.Call(inReg)
				// handle error
				if len(ret) > 0 {
					rErr := ret[0].Interface()
					if rErr != nil {
						if e, ok := rErr.(error); ok {
							c.logger.Err(e)
							return e
						} else {
							return errors.New("unknown error occurred during the function call")
						}
					}
				} else {
					return errors.New("register should return Value and error types")
				}
			}
		}
	}
	return nil
}

func (c *Cascade) traverseProvider(n *structures.DllNode, in []reflect.Value) error {
	// type implements Provider interface
	if reflect.TypeOf(n.Vertex.Iface).Implements(reflect.TypeOf((*Provider)(nil)).Elem()) {
		// if type implements Provider() it should has FnsProviderToInvoke
		if n.Vertex.Meta.FnsProviderToInvoke != nil {
			for i := 0; i < len(n.Vertex.Meta.FnsProviderToInvoke); i++ {
				m, ok := reflect.TypeOf(n.Vertex.Iface).MethodByName(n.Vertex.Meta.FnsProviderToInvoke[i])
				if !ok {
					panic("method Provides should be")
				}

				ret := m.Func.Call(in)
				// handle error
				if len(ret) > 1 {
					rErr := ret[1].Interface()
					if rErr != nil {
						if e, ok := rErr.(error); ok {
							c.logger.Err(e)
							return e
						} else {
							return errors.New("unknown error occurred during the function call")
						}
					}

					err := n.Vertex.AddValue(removePointerAsterisk(ret[0].Type().String()), ret[0], isReference(ret[0].Type()))
					if err != nil {
						return err
					}
				} else {
					return errors.New("provider should return Value and error types")
				}
			}
		}
	}
	return nil
}

func (c *Cascade) depsCall(init reflect.Method, n *structures.DllNode) error {
	in := c.getInitValues(n)

	// Iterate over dependencies
	// And search in Vertices for the provided types
	ret := init.Func.Call(in)
	rErr := ret[0].Interface()
	if rErr != nil {
		if e, ok := rErr.(error); ok {
			return e
		} else {
			return errors.New("unknown error occured during the function call")
		}
	}

	// just to be safe here
	if len(in) > 0 {
		/*
			n.Vertex.AddValue
			1. removePointerAsterisk to have uniform way of adding and searching the function args
		*/
		err := n.Vertex.AddValue(removePointerAsterisk(in[0].Type().String()), in[0], isReference(in[0].Type()))
		if err != nil {
			return err
		}
	} else {
		panic("len in less than 2")
	}

	err := c.traverseProvider(n, []reflect.Value{reflect.ValueOf(n.Vertex.Iface)})
	if err != nil {
		return err
	}

	err = c.traverseRegisters(n)
	if err != nil {
		return err
	}

	return nil
}

func (c *Cascade) getInitValues(n *structures.DllNode) []reflect.Value {
	in := make([]reflect.Value, 0, 1)

	// add service itself
	in = append(in, reflect.ValueOf(n.Vertex.Iface))

	// add dependencies
	if len(n.Vertex.Meta.InitDepsList) > 0 {
		for i := 0; i < len(n.Vertex.Meta.InitDepsList); i++ {
			depId := n.Vertex.Meta.InitDepsList[i].Name
			v := c.graph.FindProvider(depId)

			for k, val := range v.Provides {
				if k == depId {
					// value - reference and init dep also reference
					if *val.IsReference == *n.Vertex.Meta.InitDepsList[i].IsReference {
						in = append(in, *val.Value)
					} else if *val.IsReference {
						// same type, but difference in the refs
						// Init needs to be a value
						// But Vertex provided reference

						in = append(in, val.Value.Elem())
					} else if !*val.IsReference {
						// vice versa
						// Vertex provided value
						// but Init needs to be a reference
						if val.Value.CanAddr() {
							in = append(in, val.Value.Addr())
						} else {
							c.logger.Warn().Str("type", val.Value.Type().String()).Msgf("value is not addressible. TIP: consider to return a pointer from %s", val.Value.Type())
							c.logger.Warn().Msgf("making a fresh pointer")
							nt := reflect.New(val.Value.Type())
							in = append(in, nt)
						}
					}
				}
			}
		}
	}
	return in
}

func removePointerAsterisk(s string) string {
	return strings.Trim(s, "*")
}

func isReference(t reflect.Type) bool {
	return t.Kind() == reflect.Ptr
}

// TODO add all primitive types
func isPrimitive(str string) bool {
	switch str {
	case "int":
		return true
	default:
		return false
	}
}
