package endure

import (
	"reflect"

	"github.com/spiral/endure/pkg/fsm"
	"github.com/spiral/endure/pkg/vertex"
	"github.com/spiral/errors"
	"go.uber.org/zap"
)

/*
   Traverse the DLL in the forward direction

*/
func (e *Endure) internalInit(vrtx *vertex.Vertex) error {
	const op = errors.Op("endure_internal_init")
	if vrtx.IsDisabled {
		e.logger.Warn("vertex is disabled due to error.Disabled in the Init func or due to Endure decision (Disabled dependency)", zap.String("vertex id", vrtx.ID))
		return nil
	}
	// we already checked the Interface satisfaction
	// at this step absence of Init() is impoosssibruuu
	initMethod, _ := reflect.TypeOf(vrtx.Iface).MethodByName(InitMethodName)

	err := e.callInitFn(initMethod, vrtx)
	if err != nil {
		if errors.Is(errors.Disabled, err) {
			return err
		}
		e.logger.Error("error occurred during the call INIT function", zap.String("vertex id", vrtx.ID), zap.Error(err))
		return errors.E(op, errors.FunctionCall, err)
	}

	return nil
}

/*
Here we also track the Disabled vertices. If the vertex is disabled we should re-calculate the tree

*/
func (e *Endure) callInitFn(init reflect.Method, vrtx *vertex.Vertex) error {
	const op = errors.Op("endure_call_init_fn")
	if vrtx.GetState() != fsm.Initializing {
		return errors.E("vertex should be in Initializing state")
	}
	in, err := e.findInitParameters(vrtx)
	if err != nil {
		return errors.E(op, errors.FunctionCall, err)
	}
	// Iterate over dependencies
	// And search in Vertices for the provided types
	ret := init.Func.Call(in)
	rErr := ret[0].Interface()
	if rErr != nil {
		if err, ok := rErr.(error); ok && err != nil {
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
				e.logger.Warn("vertex disabled", zap.String("vertex id", vrtx.ID), zap.Error(err))
				// disable current vertex
				vrtx.IsDisabled = true
				// Disabled is actually to an error, just notification to the graph, that it has some vertices which are disabled
				return errors.E(op, errors.Disabled)
			}

			e.logger.Error("error calling internal_init", zap.String("vertex id", vrtx.ID), zap.Error(err))
			return errors.E(op, errors.FunctionCall, err)
		}
		return errors.E(op, errors.FunctionCall, errors.Str("unknown error occurred during the function call"))
	}

	// just to be safe here
	// len should be at least 1 (receiver)
	if len(in) > 0 {
		/*
			n.Vertex.AddProvider
			1. removePointerAsterisk to have uniform way of adding and searching the function args
			2. if value already exists, AddProvider will replace it with new one
		*/
		vrtx.AddProvider(removePointerAsterisk(in[0].Type().String()), in[0], isReference(in[0].Type()), in[0].Kind())
		e.graph.AddGlobalProvider(removePointerAsterisk(in[0].Type().String()), in[0])
		e.logger.Debug("value added successfully", zap.String("vertex id", vrtx.ID), zap.String("parameter", in[0].Type().String()))
	} else {
		e.logger.Error("0 or less parameters for Init", zap.String("vertex id", vrtx.ID))
		return errors.E(op, errors.ArgType, errors.Str("0 or less parameters for Init"))
	}

	if len(vrtx.Meta.CollectorEntries) > 0 {
		for i := 0; i < len(vrtx.Meta.CollectorEntries); i++ {
			// try to find nil IN args and get it from global
			for j := 0; j < len(vrtx.Meta.CollectorEntries[i].In); j++ {
				if vrtx.Meta.CollectorEntries[i].In[j].In.IsZero() {
					global, ok := e.graph.Providers[vrtx.Meta.CollectorEntries[i].In[j].Dep]
					if !ok {
						e.logger.Error("can't find in arg to Call Collects on the vertex", zap.String("vertex id", vrtx.ID))
						return errors.E(op, errors.Errorf("vertex id: %s", vrtx.ID))
					}
					vrtx.Meta.CollectorEntries[i].In[j].In = global
				}
			}

			in := make([]reflect.Value, 0, len(vrtx.Meta.CollectorEntries[i].In))
			for _, v := range vrtx.Meta.CollectorEntries[i].In {
				in = append(in, v.In)
			}
			err = e.fnCallCollectors(vrtx, in, vrtx.Meta.CollectorEntries[i].Fn)
			if err != nil {
				return errors.E(op, errors.Traverse, err)
			}
		}
	}
	return nil
}

func (e *Endure) findInitParameters(vrtx *vertex.Vertex) ([]reflect.Value, error) {
	const op = errors.Op("endure_find_init_parameters")
	in := make([]reflect.Value, 0, 2)

	// add service itself
	in = append(in, reflect.ValueOf(vrtx.Iface))

	// add dependencies
	if len(vrtx.Meta.InitDepsToInvoke) > 0 {
		for depID := range vrtx.Meta.InitDepsToInvoke {
			fnReceiver := e.graph.VerticesMap[depID]
			calleeVertexID := vrtx.ID
			err := e.traverseProviders(fnReceiver, calleeVertexID)
			if err != nil {
				return nil, errors.E(op, errors.Traverse, err)
			}
		}

		// TODO algorithm of minimum compatibility
		for _, o := range vrtx.Meta.InitDepsOrd {
			entries := vrtx.Meta.InitDepsToInvoke[o]
			for i := 0; i < len(entries); i++ {
				in = append(in, e.graph.Providers[entries[i].Name])
			}
		}
	}

	return in, nil
}
