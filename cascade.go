package cascade

import (
	"fmt"
	"reflect"
)

type Cascade struct {
	providers map[reflect.Type]entry
	registers map[reflect.Type]entry
	services  *serviceGraph
}

type entry struct {
	name string
	node interface{}
}

func NewContainer() *Cascade {
	return &Cascade{
		registers: nil,
		providers: nil,
		services:  nil,
	}
}

func (c *Cascade) Register(name string, service Service) error {
	if c.services.has(name) {
		return fmt.Errorf("service `%s` already exists", name)
	}

	c.services.push(name, service)

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

			c.registers[argsTypes[0]] = entry{name: name, node: fn}
		}
	}

	return nil
}

// Init container and all service dependencies.
func (c *Cascade) Init() error {
	if err := c.calculateDependencies(); err != nil {
		return err
	}
}

func (c *Cascade) calculateDependencies() error {
	// Calculate service dependencies
	for name, node := range c.services.nodes {
		init, ok := reflect.TypeOf(node).MethodByName("Init")
		if !ok {
			// no init method
			continue
		}

		initArgs, err := argrType(init)
		if err != nil {
			return err
		}

		for _, arg := range initArgs {
			for nn, nd := range c.services.nodes {
				if typeMatches(arg, nd) {
					// found dependency via Init method
					c.services.depends(name, nn)
				}
			}

			for t, e := range c.providers {
				if typeMatches(arg, t) {
					// found dependency via Init method (provided by Provider)
					c.services.depends(name, e.name)
				}
			}
		}
	}

	for t, e := range c.registers {
		for sn, se := range c.services.nodes {
			if typeMatches(t, se) {
				// depends via dynamic dependency declared as Registers method
				c.services.depends(e.name, sn)
			}
		}
	}
}
