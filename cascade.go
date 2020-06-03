package cascade

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/spiral/cascade/structures"
)

const Init = "Init"

type Cascade struct {
	deps []*structures.Dep

	providers     map[reflect.Type]entry
	depends       map[reflect.Type][]entry
	servicesGraph *structures.Graph
}

type entry struct {
	name   string
	vertex interface{}
}

func NewContainer() *Cascade {
	return &Cascade{
		deps:          []*structures.Dep{},
		depends:       make(map[reflect.Type][]entry),
		providers:     make(map[reflect.Type]entry),
		servicesGraph: structures.NewAL(),
	}
}

// Register depends the dependencies
// name is a name of the dependency, for example - S2
// vertex is a value -> pointer to the structure
func (c *Cascade) Register(name string, vertex interface{}) error {
	if c.servicesGraph.Has(name) {
		return fmt.Errorf("vertex `%s` already exists", name)
	}

	// just push the vertex
	// here we can append in future some meta information
	c.servicesGraph.AddVertex(name, vertex, reflect.TypeOf(vertex).String())

	if provider, ok := vertex.(Provider); ok {
		for _, fn := range provider.Provides() {
			ret, err := returnType(fn)
			if err != nil {
				// todo: delete vertex
				return err
			}
			// save providers
			c.providers[ret] = entry{name: name, vertex: fn}
		}
	}

	if register, ok := vertex.(Register); ok {
		for _, fn := range register.Depends() {
			// what type it might depend on?
			argsTypes, err := argType(fn)
			if err != nil {
				// todo: delete vertex
				return err
			}

			if len(argsTypes) != 1 {
				return fmt.Errorf("%s must accept exactly one argument", fn)
			}

			if len(argsTypes) > 0 {
				// if we found, that some structure depends on some type
				// we also save it in the `depends` section
				// name s1 (for example)
				// vertex - S4 func
				c.depends[argsTypes[0]] = append(c.depends[argsTypes[0]], entry{name: name, vertex: fn})
			} else {
				// todo temporary
				panic("argsTypes less than 0")
			}
		}
	}

	return nil
}

// Init container and all service edges.
func (c *Cascade) Init() error {
	// traverse the graph
	if err := c.calculateDependencies(); err != nil {
		return err
	}

	return nil
}

// calculateDependencies calculates simple graph for the dependencies
func (c *Cascade) calculateDependencies() error {
	// name for example S2
	for name, vrtx := range c.servicesGraph.Vertices {
		init, ok := reflect.TypeOf(vrtx.Value).MethodByName(Init)
		if !ok {
			continue
		}

		// S2 init args
		initArgs, err := functionParameters(init)
		if err != nil {
			return err
		}

		// iterate over all function parameters
		for _, initArg := range initArgs {
			for id, vertex := range c.servicesGraph.Vertices {
				if id == name {
					continue
				}

				initArgTr := removePointerAsterisk(initArg.String())
				vertexTypeTr := removePointerAsterisk(reflect.TypeOf(vertex.Value).String())

				// guess, the types are the same type
				if initArgTr == vertexTypeTr {
					c.servicesGraph.AddEdge(name, id)
				}
			}

			// provides type (DB for example)
			// and entry for that type
			for t, e := range c.providers {
				provider := removePointerAsterisk(t.String())

				if provider == initArg.String() {
					c.servicesGraph.AddEdge(name, e.name)
				}
			}
		}

		// second round of the dependencies search
		// via the depends
		// in the tests, S1 depends on the S4 and S2 on the S4 via the Depends interface
		// a lot of stupid allocations here, needed to be optimized in the future
		for rflType, slice := range c.depends {
			// check if we iterate over the needed type
			if removePointerAsterisk(rflType.String()) == removePointerAsterisk(vrtx.Meta.RawPackage) {
				for _, entry := range slice {
					// rflType --> S4
					// in slice s1, s2

					// guard here
					entryType, _ := argType(entry.vertex)
					if len(entryType) > 0 {
						if removePointerAsterisk(entryType[0].String()) == removePointerAsterisk(rflType.String()) {
							c.servicesGraph.AddEdge(entry.name, name)
						}
					}
				}
			}
		}
	}

	return nil
}

func (c *Cascade) flattenSimpleGraph() []structures.Dep {
	
	return nil
}

func removePointerAsterisk(s string) string {
	return strings.Trim(s, "*")
}
