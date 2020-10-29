package endure

import (
	"reflect"

	"github.com/spiral/endure/structures"
	"github.com/spiral/errors"
	"go.uber.org/zap"
)

/*
   Traverse the DLL in the forward direction

*/
func (e *Endure) init(vertex *structures.Vertex) error {
	const op = errors.Op("internal_init")
	if vertex.IsDisabled {
		e.logger.Warn("vertex is disabled due to error.Disabled in the Init func or due to Endure decision (Disabled dependency)", zap.String("vertex id", vertex.ID))
		return nil
	}
	// we already checked the Interface satisfaction
	// at this step absence of Init() is impoosssibruuu
	initMethod, _ := reflect.TypeOf(vertex.Iface).MethodByName(InitMethodName)

	err := e.callInitFn(initMethod, vertex)
	if err != nil {
		e.logger.Error("error occurred during the call INIT function", zap.String("vertex id", vertex.ID), zap.Error(err))
		return errors.E(op, errors.FunctionCall, err)
	}

	return nil
}

/*
Here we also track the Disabled vertices. If the vertex is disabled we should re-calculate the tree

*/
func (e *Endure) callInitFn(init reflect.Method, vertex *structures.Vertex) error {
	const op = errors.Op("internal_call_init_function")
	in, err := e.findInitParameters(vertex)
	if err != nil {
		return errors.E(op, errors.FunctionCall, err)
	}
	// Iterate over dependencies
	// And search in Vertices for the provided types
	ret := init.Func.Call(in)
	rErr := ret[0].Interface()
	if rErr != nil {
		if err, ok := rErr.(error); ok && e != nil {
			/*
				If vertex is disabled we skip all processing for it:
				1. We don't add Init function args as dependencies
			*/
			if errors.Is(errors.Disabled, err) {
				/*
					DisableById vertex
					1. But if vertex is disabled it can't PROVIDE via v.Provided value of itself for other vertices
					and we should recalculate whole three without this dep.
				*/
				e.logger.Warn("vertex disabled", zap.String("vertex id", vertex.ID), zap.Error(err))
				// disable current vertex
				vertex.IsDisabled = true
				// disable all vertices in the vertex which depends on current
				e.graph.DisableById(vertex.ID)
				// Disabled is actually to an error, just notification to the graph, that it has some vertices which are disabled
				return nil
			} else {
				e.logger.Error("error calling init", zap.String("vertex id", vertex.ID), zap.Error(err))
				return errors.E(op, errors.FunctionCall, err)
			}
		} else {
			return errors.E(op, errors.FunctionCall, errors.Str("unknown error occurred during the function call"))
		}
	}

	// just to be safe here
	// len should be at least 1 (receiver)
	if len(in) > 0 {
		/*
			n.Vertex.AddProvider
			1. removePointerAsterisk to have uniform way of adding and searching the function args
			2. if value already exists, AddProvider will replace it with new one
		*/
		vertex.AddProvider(removePointerAsterisk(in[0].Type().String()), in[0], isReference(in[0].Type()), in[0].Kind())
		e.logger.Debug("value added successfully", zap.String("vertex id", vertex.ID), zap.String("parameter", in[0].Type().String()))
	} else {
		e.logger.Error("0 or less parameters for Init", zap.String("vertex id", vertex.ID))
		return errors.E(op, errors.ArgType, errors.Str("0 or less parameters for Init"))
	}

	if len(vertex.Meta.CollectsDepsToInvoke) > 0 {
		for i := 0; i < len(vertex.Meta.CollectsDepsToInvoke); i++ {
			// Interface dependency
			if vertex.Meta.CollectsDepsToInvoke[i].Kind == reflect.Interface {
				err = e.traverseCallCollectorsInterface(vertex)
				if err != nil {
					return errors.E(op, errors.Traverse, err)
				}
			} else {
				// structure dependence
				err = e.traverseCallCollectors(vertex)
				if err != nil {
					return errors.E(op, errors.Traverse, err)
				}
			}
		}
	}
	return nil
}

func (e *Endure) findInitParameters(vertex *structures.Vertex) ([]reflect.Value, error) {
	const op = errors.Op("internal_find_init_parameters")
	in := make([]reflect.Value, 0, 2)

	// add service itself
	in = append(in, reflect.ValueOf(vertex.Iface))

	// add dependencies
	if len(vertex.Meta.InitDepsToInvoke) > 0 {
		for i := 0; i < len(vertex.Meta.InitDepsToInvoke); i++ {
			depID := vertex.Meta.InitDepsToInvoke[i].Name
			v := e.graph.FindProviders(depID)
			var err error
			in, err = e.traverseProviders(vertex.Meta.InitDepsToInvoke[i], v[0], depID, vertex.ID, in)
			if err != nil {
				return nil, errors.E(op, errors.Traverse, err)
			}
		}
	}
	return in, nil
}
