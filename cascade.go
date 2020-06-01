package cascade

import (
	"fmt"
	"reflect"

	"github.com/spiral/cascade/data_structures"
)

type Cascade struct {
	Deps []*data_structures.Dep

	providers     map[reflect.Type]entry
	depends       map[reflect.Type]entry
	servicesGraph *data_structures.Graph
}

type entry struct {
	name string
	node interface{}
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
func (c *Cascade) Register(name string, service interface{}) error {
	if c.servicesGraph.Has(name) {
		return fmt.Errorf("service `%s` already exists", name)
	}

	// just push the node
	// here we can append in future some meta information
	c.servicesGraph.AddVertex(name, service)

	if provider, ok := service.(Provider); ok {
		for _, fn := range provider.Provides() {
			ret, err := returnType(fn)
			if err != nil {
				// todo: delete service
				return err
			}

			c.providers[ret] = entry{name: name, node: fn}
		}
	}

	if register, ok := service.(Register); ok {
		for _, fn := range register.Depends() {
			// what type it might depend on?
			argsTypes, err := argType(fn)
			if err != nil {
				// todo: delete service
				return err
			}

			if len(argsTypes) != 1 {
				return fmt.Errorf("%s must accept exactly one argument", fn)
			}

			if len(argsTypes) > 0 {
				c.depends[argsTypes[0]] = entry{name: name, node: fn}
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

		init, ok := reflect.TypeOf(node.Value).MethodByName("Init")
		if !ok {
			// no init method
			continue
		}

		// get arg types from the Init methods Init(a A1, b B1)
		// A1 and B1 types will be in initArgs
		initArgs, err := functionParameters(init)
		if err != nil {
			return err
		}

		// interate over all args
		for _, arg := range initArgs {
			for verticesName, nd := range c.servicesGraph.Vertices {
				if verticesName == name {
					continue
				}

				if reflect.TypeOf(nd.Value).ConvertibleTo(arg) {
					// found dependency via Init method
					c.servicesGraph.AddEdge(name, verticesName)
				}
			}

			for reflectType, entry := range c.providers {
				if reflect.TypeOf(entry).ConvertibleTo(reflectType) {
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
		for srvName, srvNode := range c.servicesGraph.Vertices {
			if reflect.TypeOf(srvNode).ConvertibleTo(reflectType) {
				// depends via dynamic dependency declared as Depends method
				c.servicesGraph.AddEdge(entry.name, srvName)
			}
		}

		// todo: do we need it?
		for providersType, prvEntry := range c.providers {
			if reflect.TypeOf(providersType).ConvertibleTo(reflectType) {
				// found dependency via Init method (provided by Provider)
				c.servicesGraph.AddEdge(entry.name, prvEntry.name)
			}
		}
	}



	return nil
}
