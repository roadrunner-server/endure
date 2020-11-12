package endure

import (
	"reflect"

	"github.com/spiral/errors"
	"go.uber.org/zap"
)

/*
   Traverse the DLL in the forward direction

*/
func (e *Endure) internalInit(vertex *Vertex) error {
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
func (e *Endure) callInitFn(init reflect.Method, vertex *Vertex) error {
	const op = errors.Op("call init and collects")
	if vertex.GetState() != Initializing {
		return errors.E("vertex should be in Initializing state")
	}
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
				e.logger.Error("error calling internal_init", zap.String("vertex id", vertex.ID), zap.Error(err))
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
		e.graph.AddGlobalProvider(removePointerAsterisk(in[0].Type().String()), in[0])
		e.logger.Debug("value added successfully", zap.String("vertex id", vertex.ID), zap.String("parameter", in[0].Type().String()))
	} else {
		e.logger.Error("0 or less parameters for Init", zap.String("vertex id", vertex.ID))
		return errors.E(op, errors.ArgType, errors.Str("0 or less parameters for Init"))
	}

	if len(vertex.Meta.FnsCollectorToInvoke) > 0 {
		for i := 0; i < len(vertex.Meta.FnsCollectorToInvoke); i++ {
			// try to find nil IN args and get it from global
			for j := 0; j < len(vertex.Meta.FnsCollectorToInvoke[i].in); j++ {
				if vertex.Meta.FnsCollectorToInvoke[i].in[j].in.IsZero() {
					global, ok := e.graph.providers[vertex.Meta.FnsCollectorToInvoke[i].in[j].dep]
					if !ok {
						e.logger.Error("can't find in arg to Call Collects on the vertex", zap.String("vertex id", vertex.ID))
						return errors.E(op, errors.Errorf("vertex id: %s", vertex.ID))
					}
					vertex.Meta.FnsCollectorToInvoke[i].in[j].in = global
				}
			}

			in := make([]reflect.Value, 0, len(vertex.Meta.FnsCollectorToInvoke[i].in))
			for _, v := range vertex.Meta.FnsCollectorToInvoke[i].in {
				in = append(in, v.in)
			}
			err = e.callCollectorFns(vertex, in, vertex.Meta.FnsCollectorToInvoke[i].fn)
			if err != nil {
				return errors.E(op, errors.Traverse, err)
			}
		}
	}
	return nil
}

func (e *Endure) findInitParameters(vertex *Vertex) ([]reflect.Value, error) {
	const op = errors.Op("internal_find_init_parameters")
	in := make([]reflect.Value, 0, 2)

	// add service itself
	in = append(in, reflect.ValueOf(vertex.Iface))

	// add dependencies
	if len(vertex.Meta.InitDepsToInvoke) > 0 {
		for depID := range vertex.Meta.InitDepsToInvoke {
			fnReceiver := e.graph.VerticesMap[depID]
			calleeVertexId := vertex.ID
			err := e.traverseProviders(fnReceiver, calleeVertexId)
			if err != nil {
				return nil, errors.E(op, errors.Traverse, err)
			}
		}

		// TODO algorithm of minimum compatibility
		for _, o := range vertex.Meta.InitDepsOrd {
			entries := vertex.Meta.InitDepsToInvoke[o]
			for i := 0; i < len(entries); i++ {
				in = append(in, e.graph.providers[entries[i].Name])
			}
		}
	}

	return in, nil
}
