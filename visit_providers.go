package endure

import (
	"fmt"
	"reflect"

	"github.com/spiral/errors"
	"go.uber.org/zap"
)

func (e *Endure) traverseProviders(depsEntry Entry, depVertex *Vertex, depID string, calleeID string, in []reflect.Value) ([]reflect.Value, error) {
	const op = errors.Op("internal_traverse_providers")
	err := e.traverseCallProvider(depVertex, []reflect.Value{reflect.ValueOf(depVertex.Iface)}, calleeID, depID)
	if err != nil {
		return nil, errors.E(op, errors.Traverse, err)
	}

	// to index function name in defer
	for providerID, providedEntry := range depVertex.Provides {
		if providerID == depID {
			in = e.appendProviderFuncArgs(depsEntry, providedEntry, in)
		}
	}

	return in, nil
}

func (e *Endure) appendProviderFuncArgs(depsEntry Entry, providedEntry ProvidedEntry, in []reflect.Value) []reflect.Value {
	switch {
	case *providedEntry.IsReference == *depsEntry.IsReference:
		in = append(in, *providedEntry.Value)
	case *providedEntry.IsReference:
		// same type, but difference in the refs
		// Init needs to be a value
		// But Vertex provided reference
		in = append(in, providedEntry.Value.Elem())
	case !*providedEntry.IsReference:
		// vice versa
		// Vertex provided value
		// but Init needs to be a reference
		if providedEntry.Value.CanAddr() {
			in = append(in, providedEntry.Value.Addr())
		} else {
			e.logger.Warn(fmt.Sprintf("value is not addressible. TIP: consider to return a pointer from %s", providedEntry.Value.Type()), zap.String("type", providedEntry.Value.Type().String()))
			e.logger.Warn("making a fresh pointer")
			nt := reflect.New(providedEntry.Value.Type())
			in = append(in, nt)
		}
	}
	return in
}

func (e *Endure) traverseCallProvider(vertex *Vertex, in []reflect.Value, callerID, depId string) error {
	const op = errors.Op("internal_traverse_call_provider")
	// to index function name in defer
	i := 0
	defer func() {
		if r := recover(); r != nil {
			e.logger.Error("panic during the function call", zap.String("function name", vertex.Meta.FnsProviderToInvoke[i].FunctionName), zap.String("error", fmt.Sprint(r)))
		}
	}()
	// type implements Provider interface
	if reflect.TypeOf(vertex.Iface).Implements(reflect.TypeOf((*Provider)(nil)).Elem()) {
		// if type implements Provider() it should has FnsProviderToInvoke
		if vertex.Meta.FnsProviderToInvoke != nil {
			// go over all function to invoke
			// invoke it
			// and save its return values
			for i = 0; i < len(vertex.Meta.FnsProviderToInvoke); i++ {
				m, ok := reflect.TypeOf(vertex.Iface).MethodByName(vertex.Meta.FnsProviderToInvoke[i].FunctionName)
				if !ok {
					e.logger.Panic("should implement the Provider interface", zap.String("function name", vertex.Meta.FnsProviderToInvoke[i].FunctionName))
				}

				if vertex.Meta.FnsProviderToInvoke[i].ReturnTypeId != depId {
					continue
				}

				/*
				   think about better solution here TODO
				   We copy IN params here because only in slice is constant
				*/
				inCopy := make([]reflect.Value, len(in))
				copy(inCopy, in)

				/*
					cases when func NumIn can be more than one
					is that function accepts some other type except of receiver
					at the moment we assume, that this "other type" is FunctionName interface
				*/
				if m.Func.Type().NumIn() > 1 {
					/*
						here we should add type which implement Named interface
						at the moment we seek for implementation in the callerID only
					*/

					callerV := e.graph.GetVertex(callerID)
					if callerV == nil {
						return errors.E(op, errors.Traverse, errors.Str("caller vertex is nil"))
					}

					// skip function receiver
					for j := 1; j < m.Func.Type().NumIn(); j++ {
						// current function IN type (interface)
						t := m.Func.Type().In(j)
						if t.Kind() != reflect.Interface {
							e.logger.Panic("Provider accepts only interfaces", zap.String("function name", vertex.Meta.FnsProviderToInvoke[i].FunctionName))
						}

						// if Caller struct implements interface -- ok, add it to the inCopy list
						// else panic
						if reflect.TypeOf(callerV.Iface).Implements(t) == false {
							e.logger.Panic("Caller should implement callee interface", zap.String("function name", vertex.Meta.FnsProviderToInvoke[i].FunctionName))
						}

						inCopy = append(inCopy, reflect.ValueOf(callerV.Iface))
					}
				}

				ret := m.Func.Call(inCopy)
				// handle error
				if len(ret) > 1 {
					rErr := ret[1].Interface()
					if rErr != nil {
						if err, ok := rErr.(error); ok && e != nil {
							e.logger.Error("error occurred in the traverseCallProvider", zap.String("vertex id", vertex.ID))
							return errors.E(op, errors.FunctionCall, err)
						}
						return errors.E(op, errors.FunctionCall, errors.Str("unknown error occurred during the function call"))
					}

					// add the value to the Providers
					e.logger.Debug("value added successfully", zap.String("vertex id", vertex.ID), zap.String("caller id", callerID), zap.String("parameter", in[0].Type().String()))
					vertex.AddProvider(removePointerAsterisk(ret[0].Type().String()), ret[0], isReference(ret[0].Type()), in[0].Kind())
				} else {
					return errors.E(op, errors.FunctionCall, errors.Str("provider should return Value and error types"))
				}
			}
		}
	}
	return nil
}
