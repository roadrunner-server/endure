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
	graph *structures.Graph
}

func NewContainer() *Cascade {
	return &Cascade{
		graph: structures.NewGraph(),
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
		RawPackage: rawTypeStr,
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

	o := c.graph.Order()
	fmt.Println(o)

	return nil
}

// calculateEdges calculates simple graph for the dependencies
func (c *Cascade) calculateEdges() error {
	// vertexID for example S2
	for vertexID, vrtx := range c.graph.Graph {
		init, ok := reflect.TypeOf(vrtx.Value).MethodByName(Init)
		if !ok {
			panic("init method should be implemented")
		}
		_ = init

		/* Add the dependencies (if) which this vertex needs to run
		Information we know at this step is:
		1. VertexId
		2. Vertex structure value (interface)
		3. Provided type
		4. Provided type String name
		5. Name of the dependencies which we should found
		We add 3 and 4 points to the Vertex
		*/
		err := c.calculateRegisterDeps(vertexID, vrtx.Value)
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

func removePointerAsterisk(s string) string {
	return strings.Trim(s, "*")
}
