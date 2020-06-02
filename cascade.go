package cascade

import (
	"fmt"
	"reflect"

	"github.com/spiral/cascade/data_structures"
)

const Init = "Init"

type Cascade struct {
	Deps []*data_structures.Dep

	providers     map[reflect.Type]entry
	depends       map[reflect.Type]entry
	servicesGraph *data_structures.Graph
}

type entry struct {
	name   string
	vertex interface{}
}

func NewContainer() *Cascade {
	return &Cascade{
		Deps:          []*data_structures.Dep{},
		depends:       make(map[reflect.Type]entry),
		providers:     make(map[reflect.Type]entry),
		servicesGraph: data_structures.NewAL(),
	}
}

// Register depends the dependencies
func (c *Cascade) Register(name string, vertex interface{}) error {
	if c.servicesGraph.Has(name) {
		return fmt.Errorf("vertex `%s` already exists", name)
	}

	// just push the vertex
	// here we can append in future some meta information
	c.servicesGraph.AddVertex(name, vertex)

	if provider, ok := vertex.(Provider); ok {
		for _, fn := range provider.Provides() {
			ret, err := returnType(fn)
			if err != nil {
				// todo: delete vertex
				return err
			}

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
				c.depends[argsTypes[0]] = entry{name: name, vertex: fn}
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

//
func (c *Cascade) calculateDependencies() error {
	// Calculate service edges
	for name, node := range c.servicesGraph.Vertices {
		//d := &data_structures.Dep{
		//	Id: name,
		//}

		init, ok := reflect.TypeOf(node.Value).MethodByName(Init)
		if !ok {
			// no init method
			continue
		}

		// get arg types from the Init methods Init(a A1, b B1) + receiver
		// A1 and B1 types will be in initArgs
		initArgs, err := functionParameters(init)
		if err != nil {
			return err
		}

		for _, arg := range initArgs {
			for vertexName, vertex := range c.servicesGraph.Vertices {
				if vertexName == name {
					continue
				}

				if typeMatches(arg, vertex.Value) {
					c.servicesGraph.AddEdge(name, vertexName)
				}
			}
		}


		// interate over all args
		for _, arg := range initArgs {
			for vertexName, vertex := range c.servicesGraph.Vertices {
				if vertexName == name {
					continue
				}

				if typeMatches(arg, vertex.Value) {
					// found dependency via Init method
					c.servicesGraph.AddEdge(vertexName, name)
				}
			}

			for reflectType, entry := range c.providers {
				if typeMatches(reflectType, entry.vertex) {
					//d.D = e
					//c.Deps = append(c.Deps, d)
					// found dependency via Init method (provided by Provider)
					c.servicesGraph.AddEdge(name, entry.name)
				}
			}
		}
	}

	// iterate over all registered types
	for reflectType, entry := range c.depends {
		for vertexName, vertex := range c.servicesGraph.Vertices {
			if typeMatches(reflectType, vertex.Value) {
				// depends via dynamic dependency declared as Depends method
				c.servicesGraph.AddEdge(entry.name, vertexName)
			}
		}

		// todo: do we need it?
		for providersType, prvEntry := range c.providers {
			if typeMatches(providersType, prvEntry.vertex) {
				// found dependency via Init method (provided by Provider)
				c.servicesGraph.AddEdge(entry.name, prvEntry.name)
			}
		}
	}



	return nil
}
