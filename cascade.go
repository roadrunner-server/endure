package cascade

import (
	"fmt"
	"reflect"

	"github.com/spiral/cascade/data_structures"
)

type Cascade struct {
	providers map[reflect.Type]entry
	registers map[reflect.Type]entry
	services  *data_structures.Graph
}

type entry struct {
	name string
	node interface{}
}

func NewContainer() *Cascade {
	return &Cascade{
		registers: make(map[reflect.Type]entry),
		providers: make(map[reflect.Type]entry),
		services: &data_structures.Graph{
			Nodes: map[string]data_structures.Node{},
			Edges: map[string][]string{},
		},
	}
}

// Register registers the dependencies
func (c *Cascade) Register(name string, service interface{}) error {
	if c.services.Has(name) {
		return fmt.Errorf("service `%s` already exists", name)
	}

	c.services.Push(name, service)

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
		for _, fn := range register.Registers() {
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
				c.registers[argsTypes[0]] = entry{name: name, node: fn}
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
	for name, node := range c.services.Nodes {
		init, ok := reflect.TypeOf(node.Value).MethodByName("Init")
		if !ok {
			// no init method
			continue
		}

		// get arg types from the Init methods Init(a A1, b B1)
		// A1 and B1 types will be in initArgs
		initArgs, err := argrType(init)
		if err != nil {
			return err
		}

		// interate over all args 
		for _, arg := range initArgs {
			for nn, nd := range c.services.Nodes {
				if nn == name {
					continue
				}

				if typeMatches(arg, nd.Value) {
					// found dependency via Init method
					c.services.Depends(name, nn)
				}
			}

			for t, e := range c.providers {
				if typeMatches(arg, t) {
					// found dependency via Init method (provided by Provider)
					c.services.Depends(name, e.name)
				}
			}
		}
	}

	// iterate over all registered types
	for t, e := range c.registers {
		for sn, se := range c.services.Nodes {
			if typeMatches(t, se.Value) {
				// depends via dynamic dependency declared as Registers method
				c.services.Depends(e.name, sn)
			}
		}

		// todo: do we need it?
		for tp, te := range c.providers {
			if typeMatches(t, tp) {
				// found dependency via Init method (provided by Provider)
				c.services.Depends(e.name, te.name)
			}
		}
	}

	return nil
}
