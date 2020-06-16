package cascade

import (
	"errors"
	"reflect"

	"github.com/spiral/cascade/structures"
)

func (c *Cascade) depsCall(init reflect.Method, n *structures.DllNode) error {
	in := c.getInitValues(n)

	// Iterate over dependencies
	// And search in Vertices for the provided types
	ret := init.Func.Call(in)
	rErr := ret[0].Interface()
	if rErr != nil {
		if e, ok := rErr.(error); ok {
			c.logger.Err(e)
			return e
		} else {
			return unknownErrorOccurred
		}
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

	err := c.traverseProvider(n, []reflect.Value{reflect.ValueOf(n.Vertex.Iface)})
	if err != nil {
		return err
	}

	err = c.traverseRegisters(n)
	if err != nil {
		return err
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
		if e, ok := rErr.(error); ok {
			c.logger.Err(e)
			return e
		} else {
			return unknownErrorOccurred
		}
	}
	// just to be safe here
	if len(in) > 0 {
		// `in` type here is initialized function receiver
		c.logger.Info().
			Str("vertexID", n.Vertex.Id).
			Str("in parameter", in[0].Type().String()).
			Msg("calling with no deps")
		err := n.Vertex.AddValue(removePointerAsterisk(in[0].Type().String()), in[0], isReference(in[0].Type()))
		if err != nil {
			return err
		}
	}

	err := c.traverseProvider(n, in)
	if err != nil {
		c.logger.Err(err)
		return err
	}

	err = c.traverseRegisters(n)
	if err != nil {
		c.logger.Err(err)
		return err
	}

	return nil
}

func (c *Cascade) traverseRegisters(n *structures.DllNode) error {
	inReg := make([]reflect.Value, 0, 1)

	// add service itself
	inReg = append(inReg, reflect.ValueOf(n.Vertex.Iface))

	// add dependencies
	if len(n.Vertex.Meta.DepsList) > 0 {
		for i := 0; i < len(n.Vertex.Meta.DepsList); i++ {
			depId := n.Vertex.Meta.DepsList[i].Name
			v := c.graph.FindProvider(depId)

			for k, val := range v.Provides {
				if k == depId {
					// value - reference and init dep also reference
					if *val.IsReference == *n.Vertex.Meta.DepsList[i].IsReference {
						inReg = append(inReg, *val.Value)
					} else if *val.IsReference {
						// same type, but difference in the refs
						// Init needs to be a value
						// But Vertex provided reference

						inReg = append(inReg, val.Value.Elem())
					} else if !*val.IsReference {
						// vice versa
						// Vertex provided value
						// but Init needs to be a reference
						if val.Value.CanAddr() {
							inReg = append(inReg, val.Value.Addr())
						} else {
							c.logger.Warn().Str("type", val.Value.Type().String()).Msgf("value is not addressible. TIP: consider to return a pointer from %s", val.Value.Type())
							c.logger.Warn().Msgf("making a fresh pointer")

							nt := reflect.New(val.Value.Type())
							inReg = append(inReg, nt)
						}
					}
				}
			}
		}
	}

	//type implements Register interface
	if reflect.TypeOf(n.Vertex.Iface).Implements(reflect.TypeOf((*Register)(nil)).Elem()) {
		// if type implements Register() it should has FnsProviderToInvoke
		if n.Vertex.Meta.DepsList != nil {
			for i := 0; i < len(n.Vertex.Meta.FnsRegisterToInvoke); i++ {
				m, ok := reflect.TypeOf(n.Vertex.Iface).MethodByName(n.Vertex.Meta.FnsRegisterToInvoke[i])
				if !ok {
					panic("method Provides should be")
				}

				ret := m.Func.Call(inReg)
				// handle error
				if len(ret) > 0 {
					rErr := ret[0].Interface()
					if rErr != nil {
						if e, ok := rErr.(error); ok {
							c.logger.Err(e).Msg("error")
							return e
						} else {
							return unknownErrorOccurred
						}
					}
				} else {
					return errors.New("register should return Value and error types")
				}
			}
		}
	}
	return nil
}

func (c *Cascade) traverseProvider(n *structures.DllNode, in []reflect.Value) error {
	// type implements Provider interface
	if reflect.TypeOf(n.Vertex.Iface).Implements(reflect.TypeOf((*Provider)(nil)).Elem()) {
		// if type implements Provider() it should has FnsProviderToInvoke
		if n.Vertex.Meta.FnsProviderToInvoke != nil {
			for i := 0; i < len(n.Vertex.Meta.FnsProviderToInvoke); i++ {
				m, ok := reflect.TypeOf(n.Vertex.Iface).MethodByName(n.Vertex.Meta.FnsProviderToInvoke[i])
				if !ok {
					panic("method Provides should be")
				}

				ret := m.Func.Call(in)
				// handle error
				if len(ret) > 1 {
					rErr := ret[1].Interface()
					if rErr != nil {
						if e, ok := rErr.(error); ok {
							c.logger.Err(e)
							return e
						} else {
							return unknownErrorOccurred
						}
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

func (c *Cascade) getInitValues(n *structures.DllNode) []reflect.Value {
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
					} else if !*val.IsReference {
						// vice versa
						// Vertex provided value
						// but Init needs to be a reference
						if val.Value.CanAddr() {
							in = append(in, val.Value.Addr())
						} else {
							c.logger.Warn().Str("type", val.Value.Type().String()).Msgf("value is not addressible. TIP: consider to return a pointer from %s", val.Value.Type())
							c.logger.Warn().Msgf("making a fresh pointer")
							nt := reflect.New(val.Value.Type())
							in = append(in, nt)
						}
					}
				}
			}
		}
	}
	return in
}