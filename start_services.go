package cascade

import (
	"errors"
	"fmt"
	"reflect"

	"github.com/spiral/cascade/structures"
	"go.uber.org/zap"
)

func (c *Cascade) initCall(init reflect.Method, v *structures.Vertex) error {
	in := c.findInitParameters(v)

	// Iterate over dependencies
	// And search in Vertices for the provided types
	ret := init.Func.Call(in)
	rErr := ret[0].Interface()
	if rErr != nil {
		if e, ok := rErr.(error); ok && e != nil {
			c.logger.Error("error calling init", zap.String("vertex id", v.Id), zap.Error(e))
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
			2. if value already exists, AddValue will replace it with new one
		*/
		err := v.AddValue(removePointerAsterisk(in[0].Type().String()), in[0], isReference(in[0].Type()))
		if err != nil {
			return err
		}
		c.logger.Debug("value added successfully", zap.String("vertex id", v.Id), zap.String("parameter", in[0].Type().String()))

	} else {
		panic("len in less than 2")
	}

	err := c.traverseCallProvider(v, []reflect.Value{reflect.ValueOf(v.Iface)})
	if err != nil {
		return err
	}

	err = c.traverseCallRegisters(v)
	if err != nil {
		return err
	}

	return nil
}

func (c *Cascade) traverseCallRegisters(vertex *structures.Vertex) error {
	inReg := make([]reflect.Value, 0, 1)

	// add service itself
	inReg = append(inReg, reflect.ValueOf(vertex.Iface))

	// add dependencies
	if len(vertex.Meta.DepsList) > 0 {
		for i := 0; i < len(vertex.Meta.DepsList); i++ {
			depId := vertex.Meta.DepsList[i].Name
			v := c.graph.FindProvider(depId)

			for k, val := range v.Provides {
				if k == depId {
					// value - reference and init dep also reference
					if *val.IsReference == *vertex.Meta.DepsList[i].IsReference {
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
							c.logger.Warn(fmt.Sprintf("value is not addressible. TIP: consider to return a pointer from %s", val.Value.Type()), zap.String("type", val.Value.Type().String()))
							c.logger.Warn("making a fresh pointer")

							nt := reflect.New(val.Value.Type())
							inReg = append(inReg, nt)
						}
					}
				}
			}
		}
	}

	//type implements Register interface
	if reflect.TypeOf(vertex.Iface).Implements(reflect.TypeOf((*Register)(nil)).Elem()) {
		// if type implements Register() it should has FnsProviderToInvoke
		if vertex.Meta.DepsList != nil {
			for i := 0; i < len(vertex.Meta.FnsRegisterToInvoke); i++ {
				m, ok := reflect.TypeOf(vertex.Iface).MethodByName(vertex.Meta.FnsRegisterToInvoke[i])
				if !ok {
					panic("method Provides should be")
				}

				ret := m.Func.Call(inReg)
				// handle error
				if len(ret) > 0 {
					rErr := ret[0].Interface()
					if rErr != nil {
						if e, ok := rErr.(error); ok && e != nil {
							c.logger.Error("error calling Registers", zap.String("vertex id", vertex.Id), zap.Error(e))
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

func (c *Cascade) findInitParameters(vertex *structures.Vertex) []reflect.Value {
	in := make([]reflect.Value, 0, 1)

	// add service itself
	in = append(in, reflect.ValueOf(vertex.Iface))

	// add dependencies
	if len(vertex.Meta.InitDepsList) > 0 {
		for i := 0; i < len(vertex.Meta.InitDepsList); i++ {
			depId := vertex.Meta.InitDepsList[i].Name
			v := c.graph.FindProvider(depId)

			for k, val := range v.Provides {
				if k == depId {
					// value - reference and init dep also reference
					if *val.IsReference == *vertex.Meta.InitDepsList[i].IsReference {
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
							c.logger.Warn(fmt.Sprintf("value is not addressible. TIP: consider to return a pointer from %s", val.Value.Type()), zap.String("type", val.Value.Type().String()))
							c.logger.Warn("making a fresh pointer")
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

func (c *Cascade) traverseCallProvider(v *structures.Vertex, in []reflect.Value) error {
	// type implements Provider interface
	if reflect.TypeOf(v.Iface).Implements(reflect.TypeOf((*Provider)(nil)).Elem()) {
		// if type implements Provider() it should has FnsProviderToInvoke
		if v.Meta.FnsProviderToInvoke != nil {
			for i := 0; i < len(v.Meta.FnsProviderToInvoke); i++ {
				m, ok := reflect.TypeOf(v.Iface).MethodByName(v.Meta.FnsProviderToInvoke[i])
				if !ok {
					panic("method Provides should be")
				}

				ret := m.Func.Call(in)
				// handle error
				if len(ret) > 1 {
					rErr := ret[1].Interface()
					if rErr != nil {
						if e, ok := rErr.(error); ok && e != nil {
							c.logger.Error("error occurred in the traverseCallProvider", zap.String("vertex id", v.Id))
							return e
						} else {
							return unknownErrorOccurred
						}
					}

					err := v.AddValue(removePointerAsterisk(ret[0].Type().String()), ret[0], isReference(ret[0].Type()))
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



/*
Algorithm is the following (all steps executing in the topological order):
1. Call Configure() on all services -- OPTIONAL
2. Call Serve() on all services --     MUST
3. Call Stop() on all services --      MUST
4. Call Clear() on a services, which implements this interface -- OPTIONAL
*/
// call configure on the node

//func (c *Cascade) serveVertex(v *structures.Vertex) *result {
//	nCopy := v
//	// handle all configure
//	in := make([]reflect.Value, 0, 1)
//	// add service itself
//	in = append(in, reflect.ValueOf(nCopy.Iface))
//
//
//	// call internalServe
//	//userResultsCh = append(userResultsCh, c.call(nCopy, in, ServeMethodName))
//
//
//	return nil
//}

func (c *Cascade) internalServe(v *structures.Vertex, in []reflect.Value) *result {
	m, _ := reflect.TypeOf(v.Iface).MethodByName(ServeMethodName)
	ret := m.Func.Call(in)
	res := ret[0].Interface()
	if res != nil {
		if e, ok := res.(chan error); ok && e != nil {
			return &result{
				errCh:    e,
				exit:     make(chan struct{}, 2),
				vertexId: v.Id,
			}
		}
	}
	// error, result should not be nil
	// the only one reason to be nil is to vertex return parameter (channel) is not initialized
	return nil
}

func (c *Cascade) internalConfigure(n *structures.Vertex, in []reflect.Value) error {
	m, _ := reflect.TypeOf(n.Iface).MethodByName(ConfigureMethodName)
	ret := m.Func.Call(in)
	res := ret[0].Interface()
	if res != nil {
		if e, ok := res.(error); ok && e != nil {
			return e
		}
		return unknownErrorOccurred
	}
	return nil
}

func (c *Cascade) internalStop(vId string) error {
	v := c.graph.GetVertex(vId)

	in := make([]reflect.Value, 0, 1)
	// add service itself
	in = append(in, reflect.ValueOf(v.Iface))

	err := c.stop(v.Id, in)
	if err != nil {
		c.logger.Error("error occurred during the stop", zap.String("vertex id", v.Id))
	}

	if reflect.TypeOf(v.Iface).Implements(reflect.TypeOf((*Graceful)(nil)).Elem()) {
		err = c.close(v.Id, in)
		if err != nil {
			c.logger.Error("error occurred during the close", zap.String("vertex id", v.Id))
		}
	}

	return nil
}

func (c *Cascade) stop(vId string, in []reflect.Value) error {
	v := c.graph.GetVertex(vId)
	// Call Stop() method, which returns only error (or nil)
	m, _ := reflect.TypeOf(v.Iface).MethodByName(StopMethodName)
	ret := m.Func.Call(in)
	rErr := ret[0].Interface()
	if rErr != nil {
		if e, ok := rErr.(error); ok && e != nil {
			return e
		} else {
			return unknownErrorOccurred
		}
	}
	return nil
}

// TODO add stack to the all of the log events
func (c *Cascade) close(vId string, in []reflect.Value) error {
	v := c.graph.GetVertex(vId)
	// Call Close() method, which returns only error (or nil)
	m, _ := reflect.TypeOf(v.Iface).MethodByName(CloseMethodName)
	ret := m.Func.Call(in)
	rErr := ret[0].Interface()
	if rErr != nil {
		if e, ok := rErr.(error); ok && e != nil {
			return e
		} else {
			return unknownErrorOccurred
		}
	}
	return nil
}
