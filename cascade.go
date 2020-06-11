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
	// new
	graph *structures.Graph
	// map[string]map[string][]reflect.Value
	// example Vertex S2, S2.Init() + S2.createDB
	// vertex S1, dependency S2+S2.createDB
	providers map[string]reflect.Value
	// old

	//deps map[string][]structures.Dep

	depsGraph []*structures.Vertex

	//depends       map[reflect.Type][]entry
	//servicesGraph *structures.Graph
}

//type entry struct {
//	name   string
//	vertex interface{}
//}

func NewContainer() *Cascade {
	return &Cascade{
		graph: structures.NewGraph(),
		//deps:  make(map[string][]structures.Dep),
		//depends:       make(map[reflect.Type][]entry),
		providers: make(map[string]reflect.Value),
		//servicesGraph: structures.NewGraph(),
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
			// save providers
			// put into the map with deps

			// ret foo2.DB

			//c.providers[reflect.ValueOf(ret).String()] = reflect.ValueOf(ret) //entry{name: vertexID, vertex: fn}
			// c.providers[ret] = entry{name: vertexID, vertex: fn}
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

	//c.flattenSimpleGraph()
	//s := c.topologicalSort()
	//fmt.Println(s)
	//
	//c.validateSorting(s, nil, c.depsGraph)

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

		//err = c.calculateInitEdges(vertexID, init)
		//if err != nil {
		//	return err
		//}
		//
		//err = c.calculateDepEdges(vertexID, vrtx.Meta)
		//if err != nil {
		//	return err
		//}

	}

	return nil
}

/*


 */
func (c *Cascade) calculateRegisterDeps(vertexID string, vertex interface{}) error {
	//for i := 0; i < len(c.graph.Vertices); i++ {
	//	vrtx := c.graph.Vertices[i]
	//
	//
	//}
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

			if len(argsTypes) > 0 {
				for _, at := range argsTypes {
					a := at.String()
					_ = a
					// if we found, that some structure depends on some type
					// we also save it in the `depends` section
					// name s1 (for example)
					// vertex - S4 func

					// from --> to
					c.graph.AddDep(vertexID, at.String())
					//c.depends[at] = append(c.depends[at], entry{name: vertexID, vertex: fn})
				}
			} else {
				// todo temporary
				panic("argsTypes less than 0")
			}
		}
	}

	return nil
}

func (c *Cascade) calculateInitEdges(vertexID string, method reflect.Method) error {
	// S2 init args
	initArgs, err := functionParameters(method)
	if err != nil {
		return err
	}

	// iterate over all function parameters
	for _, initArg := range initArgs {
		for id, vertex := range c.graph.Graph {
			if id == vertexID {
				continue
			}

			initArgTr := removePointerAsterisk(initArg.String())
			vertexTypeTr := removePointerAsterisk(reflect.TypeOf(vertex.Value).String())

			// guess, the types are the same type
			if initArgTr == vertexTypeTr {
				c.graph.AddEdge(vertexID, id)
			}
		}

		// think about optimization
		c.calculateProvidersEdges(vertexID, initArg)
	}
	return nil
}

func (c *Cascade) calculateProvidersEdges(vertexID string, initArg reflect.Type) {
	// provides type (DB for example)
	// and entry for that type
	// for t, e := range c.providers {
	// 	provider := removePointerAsterisk(t.String())

	// 	if provider == initArg.String() {
	// 		if c.servicesGraph.Has(vertexID) == false {
	// 			c.servicesGraph.AddEdge(vertexID, e.name)
	// 		}
	// 	}
	// }
}

func (c *Cascade) calculateDepEdges(vertexID string, meta structures.Meta) error {
	// second round of the dependencies search
	// via the depends
	// in the tests, S1 depends on the S4 and S2 on the S4 via the Depends interface
	// a lot of stupid allocations here, needed to be optimized in the future
	//for rflType, slice := range c.depends {
	//	// check if we iterate over the needed type
	//	if removePointerAsterisk(rflType.String()) == removePointerAsterisk(meta.RawPackage) {
	//		for _, entry := range slice {
	//			// rflType --> S4
	//			// in slice s1, s2
	//
	//			// guard here
	//			entryType, err := argType(entry.vertex)
	//			if err != nil {
	//				return err
	//			}
	//			if len(entryType) > 0 {
	//				for _, et := range entryType {
	//					// s3:[s4 s2 s2] TODO
	//					if removePointerAsterisk(et.String()) == removePointerAsterisk(rflType.String()) {
	//						c.servicesGraph.AddEdge(entry.name, vertexID)
	//					}
	//				}
	//			}
	//		}
	//	}
	//}
	return nil
}

func removePointerAsterisk(s string) string {
	return strings.Trim(s, "*")
}
