package cascade

import (
	"fmt"
	"reflect"
	"strings"

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
// name is a name of the dependency, for example - S2
// vertex is a value -> pointer to the structure
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
			c.providers[ret] = entry{name: ret.String(), vertex: fn}
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

				ff := reflect.TypeOf(vertex.Value)
				fn := ff.String()
				fn = strings.Trim(fn, "*")

				inittr := strings.Trim(initArg.String(), "*")

				if inittr == fn {
					c.servicesGraph.AddEdge(name, id)
				}

				// if initArgs is a vertex.Value (s2 for example)
				//if typeMatches(initArg, vertex.Value) {
				//	c.servicesGraph.AddEdge(name, id)
				//}
			}

			// provides type (DB for example)
			// and entry for that type
			for t, e := range c.providers {
				a := t.String()
				a = strings.Trim(a, "*")

				if a == initArg.String() {
					c.servicesGraph.AddEdge(name, e.name)
				}
			}
		}
	}

	return nil
}
