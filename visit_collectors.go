package endure

import (
	"fmt"
	"reflect"

	"github.com/spiral/errors"
	"go.uber.org/zap"
)

func (e *Endure) traverseCallCollectorsInterface(vertex *Vertex) error {
	const op = errors.Op("internal_traverse_call_collectors_interface")
	for i := 0; i < len(vertex.Meta.CollectsDepsToInvoke); i++ {
		// get dependency id (vertex id)
		depID := vertex.Meta.CollectsDepsToInvoke[i].Name
		// find vertex which provides dependency
		providers := e.graph.FindProviders(depID)

		// Depend from interface
		/*
			In this case we need to be careful with IN parameters
			1. We need to find type, which implements that interface
			2. Calculate IN args
			3. And invoke
		*/

		// search for providers
		for j := 0; j < len(providers); j++ {
			// vertexKey is for example foo.DB
			// vertexValue is value for that key
			for vertexKey, vertexVal := range providers[j].Provides {
				if depID != vertexKey {
					continue
				}
				// internal_init
				inInterface := make([]reflect.Value, 0, 2)
				// add service itself
				inInterface = append(inInterface, reflect.ValueOf(vertex.Iface))
				// if type provides needed type
				// value - reference and internal_init dep also reference
				switch {
				case *vertexVal.IsReference == *vertex.Meta.CollectsDepsToInvoke[i].IsReference:
					inInterface = append(inInterface, *vertexVal.Value)
				case *vertexVal.IsReference:
					// same type, but difference in the refs
					// Init needs to be a value
					// But Vertex provided reference
					inInterface = append(inInterface, vertexVal.Value.Elem())
				case !*vertexVal.IsReference:
					// vice versa
					// Vertex provided value
					// but Init needs to be a reference
					if vertexVal.Value.CanAddr() {
						inInterface = append(inInterface, vertexVal.Value.Addr())
					} else {
						e.logger.Warn(fmt.Sprintf("value is not addressible. TIP: consider to return a pointer from %s", vertexVal.Value.Type()), zap.String("type", vertexVal.Value.Type().String()))
						e.logger.Warn("making a fresh pointer")
						nt := reflect.New(vertexVal.Value.Type())
						inInterface = append(inInterface, nt)
					}
				}

				err := e.callCollectorFns(vertex, inInterface)
				if err != nil {
					return errors.E(op, errors.Traverse, err)
				}
			}
		}
	}

	return nil
}

func (e *Endure) traverseCallCollectors(vertex *Vertex) error {
	const op = "internal_traverse_call_collectors"
	in := make([]reflect.Value, 0, 2)
	// add service itself
	in = append(in, reflect.ValueOf(vertex.Iface))

	for i := 0; i < len(vertex.Meta.CollectsDepsToInvoke); i++ {
		// get dependency id (vertex id)
		depID := vertex.Meta.CollectsDepsToInvoke[i].Name
		// find vertex which provides dependency
		providers := e.graph.FindProviders(depID)
		// search for providers
		for j := 0; j < len(providers); j++ {
			for vertexID, val := range providers[j].Provides {
				// if type provides needed type
				if vertexID == depID {
					switch {
					case *val.IsReference == *vertex.Meta.CollectsDepsToInvoke[i].IsReference:
						in = append(in, *val.Value)
					case *val.IsReference:
						// same type, but difference in the refs
						// Init needs to be a value
						// But Vertex provided reference
						in = append(in, val.Value.Elem())
					case !*val.IsReference:
						// vice versa
						// Vertex provided value
						// but Init needs to be a reference
						if val.Value.CanAddr() {
							in = append(in, val.Value.Addr())
						} else {
							e.logger.Warn(fmt.Sprintf("value is not addressible. TIP: consider to return a pointer from %s", val.Value.Type()), zap.String("type", val.Value.Type().String()))
							e.logger.Warn("making a fresh pointer")
							nt := reflect.New(val.Value.Type())
							in = append(in, nt)
						}
					}
				}
			}
		}
	}

	err := e.callCollectorFns(vertex, in)
	if err != nil {
		return errors.E(op, errors.Traverse, err)
	}

	return nil
}

func (e *Endure) callCollectorFns(vertex *Vertex, in []reflect.Value) error {
	const op = errors.Op("internal_call_collector_functions")
	// type implements Collector interface
	if reflect.TypeOf(vertex.Iface).Implements(reflect.TypeOf((*Collector)(nil)).Elem()) {
		// if type implements Collector() it should has FnsProviderToInvoke
		if vertex.Meta.CollectsDepsToInvoke != nil {
			for k := 0; k < len(vertex.Meta.FnsCollectorToInvoke); k++ {
				m, ok := reflect.TypeOf(vertex.Iface).MethodByName(vertex.Meta.FnsCollectorToInvoke[k])
				if !ok {
					e.logger.Error("type has missing method in FnsCollectorToInvoke", zap.String("vertex id", vertex.ID), zap.String("method", vertex.Meta.FnsCollectorToInvoke[k]))
					return errors.E(op, errors.FunctionCall, errors.Str("type has missing method in FnsCollectorToInvoke"))
				}

				ret := m.Func.Call(in)
				// handle error
				if len(ret) > 0 {
					// error is the last return parameter in line
					rErr := ret[len(ret)-1].Interface()
					if rErr != nil {
						if err, ok := rErr.(error); ok && e != nil {
							e.logger.Error("error calling CollectorFns", zap.String("vertex id", vertex.ID), zap.Error(err))
							return errors.E(op, errors.FunctionCall, err)
						}
						return errors.E(op, errors.FunctionCall, errors.Str("unknown error occurred during the function call"))
					}
				} else {
					return errors.E(op, errors.FunctionCall, errors.Str("collector should return Value and error types"))
				}
			}
		}
	}
	return nil
}
