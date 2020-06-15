package cascade

import (
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/spiral/cascade/structures"
)

// InitMethodName is the function name for the reflection
const InitMethodName = "Init"
const Provides = "Provides"

type Cascade struct {
	// Dependency graph
	graph *structures.Graph
	// DLL used as run list to run in order
	runList *structures.DoublyLinkedList
}

func NewContainer() *Cascade {
	return &Cascade{
		graph:   structures.NewGraph(),
		runList: structures.NewDoublyLinkedList(),
	}
}

// Register depends the dependencies
// name is a name of the dependency, for example - S2
// vertex is a value -> pointer to the structure
func (c *Cascade) Register(vertex interface{}) error {
	vertexID := removePointerAsterisk(reflect.TypeOf(vertex).String())
	// Meta information
	rawTypeStr := reflect.TypeOf(vertex).String()

	meta := structures.Meta{
		RawTypeName: rawTypeStr,
	}

	/* Register the type
	Information we know at this step is:
	1. VertexId
	2. Vertex structure value (interface)
	And we fill vertex with this information
	*/
	err := c.register(vertexID, vertex, meta)
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

	return nil
}

func (c *Cascade) register(name string, vertex interface{}, meta structures.Meta) error {
	// check the vertex
	if c.graph.HasVertex(name) {
		return fmt.Errorf("vertex `%s` already exists", name)
	}

	// just push the vertex
	// here we can append in future some meta information
	c.graph.AddVertex(name, vertex, meta)
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

			if gVertex.Meta.FnsToInvoke == nil {
				gVertex.Meta.FnsToInvoke = make([]string, 0, 5)
			}

			gVertex.Meta.FnsToInvoke = append(gVertex.Meta.FnsToInvoke, functionName(fn))

			gVertex.Provides[typeStr] = structures.ProvidedEntry{
				IsReference: nil,
				Value:       nil,
			}
		}
	}
	return nil
}

// Init container and all service edges.
func (c *Cascade) Init() error {
	// traverse the graph
	if err := c.calculateEdges(); err != nil {
		return err
	}

	// we should buld runForward list in the reverse order
	// TODO return cycle error
	sortedVertices := c.graph.TopologicalSort()

	// TODO properly handle the len of the sorted vertices
	c.runList.SetHead(&structures.DllNode{
		Vertex: sortedVertices[0]})

	// TODO what if sortedVertices will contain only 1 node (len(sortedVertices) - 2 will panic)
	for i := 1; i < len(sortedVertices); i++ {
		c.runList.Push(sortedVertices[i])
	}

	return c.runForward(c.runList.Head)
}

// calculateEdges calculates simple graph for the dependencies
func (c *Cascade) calculateEdges() error {
	// vertexID for example S2
	for vertexID, vrtx := range c.graph.Graph {
		init, ok := reflect.TypeOf(vrtx.Iface).MethodByName(InitMethodName)
		if !ok {
			panic("init method should be implemented")
		}

		/* Add the dependencies (if) which this vertex needs to runForward
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

/*


 */
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
				}
			} else {
				// todo temporary
				panic("argsTypes less than 0")
			}
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
	}
	return nil
}

/*
Traverse the DLL in the forward direction

*/
func (c *Cascade) runForward(n *structures.DllNode) error {
	// traverse the dll
	for n != nil {
		init, ok := reflect.TypeOf(n.Vertex.Iface).MethodByName(InitMethodName)
		if !ok {
			panic("init method should be implemented")
		}

		initArgs, err := functionParameters(init)
		if err != nil {
			return err
		}

		// If len(initArgs) is eq to 1, than we deal with empty Init() method
		//
		if len(initArgs) == 1 {
			err = c.noDepsCall(init, n)
			if err != nil {
				return err
			}
		} else {
			// else, we deal with variadic len of Init function parameters Init(a,b,c, etc)
			// we should resolve all it all
			err = c.depsCall(init, n)
			if err != nil {
				return err
			}
		}

		// next DLL node
		n = n.Next
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
		e := rErr.(error)
		panic(e)
	}

	// just to be safe here
	if len(in) > 0 {
		err := n.Vertex.AddValue(removePointerAsterisk(in[0].Type().String()), in[0], isReference(in[0].Type()))
		if err != nil {
			return err
		}
	}

	// type implements Provider interface
	if reflect.TypeOf(n.Vertex.Iface).Implements(reflect.TypeOf((*Provider)(nil)).Elem()) {
		// if type implements Provider() it should has FnsToInvoke
		if n.Vertex.Meta.FnsToInvoke != nil {
			for i := 0; i < len(n.Vertex.Meta.FnsToInvoke); i++ {
				m, ok := reflect.TypeOf(n.Vertex.Iface).MethodByName(n.Vertex.Meta.FnsToInvoke[i])
				if !ok {
					panic("method Provides should be")
				}

				ret := m.Func.Call(in)
				// handle error
				if len(ret) > 1 {
					rErr := ret[1].Interface()
					if rErr != nil {
						e := rErr.(error)
						panic(e)
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
						//panic("choo choooooo")
					} else if !*val.IsReference {
						// vice versa
						// Vertex provided value
						// but Init needs to be a reference
						//in = append(in, *val.Value)
						panic("choo chooooooo 2")
					}
				}
			}
		}
	}

	// Iterate over dependencies
	// And search in Vertices for the provided types

	ret := init.Func.Call(in)
	rErr := ret[0].Interface()
	if rErr != nil {
		e := rErr.(error)
		panic(e)
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

	// type implements Provider interface
	if reflect.TypeOf(n.Vertex.Iface).Implements(reflect.TypeOf((*Provider)(nil)).Elem()) {
		// if type implements Provider() it should has FnsToInvoke
		if n.Vertex.Meta.FnsToInvoke != nil {
			for i := 0; i < len(n.Vertex.Meta.FnsToInvoke); i++ {
				m, ok := reflect.TypeOf(n.Vertex.Iface).MethodByName(n.Vertex.Meta.FnsToInvoke[i])
				if !ok {
					panic("method Provides should be")
				}

				ret := m.Func.Call([]reflect.Value{reflect.ValueOf(n.Vertex.Iface)})
				// handle error
				if len(ret) > 1 {
					rErr := ret[1].Interface()
					if rErr != nil {
						e := rErr.(error)
						panic(e)
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

func removePointerAsterisk(s string) string {
	return strings.Trim(s, "*")
}

func isReference(t reflect.Type) bool {
	return t.Kind() == reflect.Ptr
}

func isPrimitive(str string) bool {
	switch str {
	case "int":
		return true
	default:
		return false
	}
}
