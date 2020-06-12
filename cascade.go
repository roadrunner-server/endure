package cascade

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/spiral/cascade/structures"
)

// Init is the function name for the reflection
const Init = "Init"

type Cascade struct {
	graph   *structures.Graph
	runList *structures.DoublyLinkedList
	// concrete types
	// for example
	// foo2.DB with value
	// foo2.S2 structure
	provides map[string]*reflect.Value
	//logger za
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
				// todo: delete vertex
				return err
			}

			typeStr := ret.String()
			// get the vertex
			vertex := c.graph.GetVertex(vertexID)
			if vertex.Provides == nil {
				vertex.Provides = make(map[string]*reflect.Value)
			}
			if vertex.Provides[typeStr] == nil {
				vertex.Provides[typeStr] = &reflect.Value{}
			}

			tmp := reflect.ValueOf(ret)
			vertex.Provides[typeStr] = &tmp
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
		init, ok := reflect.TypeOf(vrtx.Iface).MethodByName(Init)
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

					// from --> to
					c.graph.AddDep(vertexID, atStr)
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

		if vertexID == removePointerAsterisk(initArg.String()) {
			continue
		}

		c.graph.AddDep(vertexID, removePointerAsterisk(initArg.String()))

	}
	return nil
}

/*
Traverse the DLL in the forward direction

*/
func (c *Cascade) runForward(n *structures.DllNode) error {
	// traverse the dll
	for n != nil {
		//println(n.Vertex.Id)

		in := make([]reflect.Value, 0, 1)

		init, ok := reflect.TypeOf(n.Vertex.Iface).MethodByName(Init)
		if !ok {
			panic("init method should be implemented")
		}

		initArgs, err := functionParameters(init)
		if err != nil {
			return err
		}

		// only service itself
		if len(initArgs) == 1 {
			for i := 0; i < init.Type.NumIn(); i++ {
				v := init.Type.In(i)

				if v.ConvertibleTo(reflect.ValueOf(n.Vertex.Iface).Type()) == true {
					in = append(in, reflect.ValueOf(n.Vertex.Iface))
				}

			}

			ret := init.Func.Call(in)
			rErr := ret[0].Interface()
			if rErr != nil {
				e := rErr.(error)
				panic(e)
			}
		}

		for i := 1; i < len(initArgs); i++ {
			aaa := initArgs[i].String()
			_ = aaa
		}

		n = n.Next
	}

	return nil
}

func removePointerAsterisk(s string) string {
	return strings.Trim(s, "*")
}

func isPrimitive(str string) bool {
	switch str {
	case "int":
		return true
	default:
		return false
	}
}
