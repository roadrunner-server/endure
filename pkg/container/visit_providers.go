package endure

import (
	"reflect"
	"sort"

	"github.com/roadrunner-server/endure/pkg/vertex"
	"github.com/roadrunner-server/errors"
	"go.uber.org/zap"
)

func (e *Endure) traverseProviders(fnReceiver *vertex.Vertex, calleeVertexID string) error {
	const op = errors.Op("endure_traverse_providers")
	err := e.traverseCallProvider(fnReceiver, []reflect.Value{reflect.ValueOf(fnReceiver.Iface)}, calleeVertexID)
	if err != nil {
		return errors.E(op, errors.Traverse, err)
	}

	return nil
}

// Providers is vertex provides type alias
type Providers []Provide

// Provide struct represents a single Provide value, which consists of:
// 1. m -> method reflected value
// 2. In -> In types (fn foo(a int, b int))
// 3. Out -> returning types (fn boo() a int, b int)
type Provide struct {
	m   reflect.Method
	In  []reflect.Type
	Out []reflect.Type
}

func (p Providers) Len() int {
	return len(p)
}

func (p Providers) Less(i, j int) bool {
	return len(p[i].In) > len(p[j].In)
}

func (p Providers) Swap(i, j int) {
	p[i].In, p[j].In = p[j].In, p[i].In
	p[i].Out, p[j].Out = p[j].Out, p[i].Out
	p[i].m, p[j].m = p[j].m, p[i].m
}

func (e *Endure) traverseCallProvider(fnReceiver *vertex.Vertex, in []reflect.Value, callerID string) error {
	const op = errors.Op("endure_traverse_call_provider")
	callerV := e.graph.GetVertex(callerID)
	if callerV == nil {
		return errors.E(op, errors.Traverse, errors.Str("caller fnReceiver is nil"))
	}

	// type implements Provider interface
	if reflect.TypeOf(fnReceiver.Iface).Implements(reflect.TypeOf((*Provider)(nil)).Elem()) {
		// if type implements Provider() it should has FnsProviderToInvoke
		if fnReceiver.Meta.FnsProviderToInvoke != nil {
			// go over all function to invoke
			// invoke it
			// and save its return values
			fnsToCall := fnReceiver.Meta.FnsProviderToInvoke.Merge()
			for i := 0; i < len(fnsToCall); i++ {
				providers := make(Providers, 0)
				for ii := 0; ii < len(fnsToCall[i]); ii++ {
					p := Provide{}
					m, ok := reflect.TypeOf(fnReceiver.Iface).MethodByName(fnsToCall[i][ii])
					if !ok {
						e.logger.Panic("should implement the Provider interface", zap.String("function fn", fnsToCall[i][ii]))
					}

					// assign current method
					p.m = m

					// example ProvideWithName(named endure.Named) (SuperInterface, error)
					// IN -> endure.Named + struct receiver
					// OUT -> SuperInterface, error
					for f := 0; f < m.Func.Type().NumIn(); f++ {
						p.In = append(p.In, m.Func.Type().In(f))
					}

					// collect out args
					// TODO will be used later, in more intellectual merge
					for ot := 0; ot < m.Func.Type().NumOut(); ot++ {
						// skip error type, record only out type
						p.Out = append(p.Out, m.Func.Type().Out(ot))
					}

					providers = append(providers, p)
				}

				// sort providers, so we will have Provider with most dependencies first
				sort.Sort(providers)

				// we know, that we  have FnsProviderToInvoke not nil here
				// we need to compare args
				for k := 0; k < len(providers); k++ {
					pr := providers[k]
					inCopy := make([]reflect.Value, len(in))
					copy(inCopy, in)

					// fallback call provided, only 1 IN arg, function receiver
					if len(pr.In) == 1 {
						err := e.fnProvidersCall(pr.m, inCopy, fnReceiver, callerID)
						if err != nil {
							return err
						}
						continue
					}

					// if we have minimum 2 In args (self and Named for example)
					// we should check where is function receiver and check if caller implement all other args
					// if everything ok we just pass first args as the receiver and caller as all the rest args
					// start from 1, 0-th index is function receiver
					// check if caller implements all needed interfaces to call func
					if !e.walk(pr.In, callerV) {
						// if not, check for other provider
						continue
					}

					for l := 1; l < len(pr.In); l++ {
						switch pr.In[l].Kind() { //nolint:exhaustive
						case reflect.Struct: // just structure
							inCopy = append(inCopy, e.graph.Providers[pr.In[l].String()])
						case reflect.Ptr: // Ptr to the structure
							val := pr.In[l].Elem() // get real value
							inCopy = append(inCopy, e.graph.Providers[val.String()])
						case reflect.Interface: // we know here, that caller implement all needed to call interfaces
							inCopy = append(inCopy, reflect.ValueOf(e.graph.VerticesMap[callerID].Iface))
						}
					}

					err := e.fnProvidersCall(pr.m, inCopy, fnReceiver, callerID)
					if err != nil {
						return err
					}
				}
			}
		}
		return nil
	}
	return nil
}

func (e *Endure) fnProvidersCall(f reflect.Method, in []reflect.Value, vertex *vertex.Vertex, callerID string) error {
	const op = errors.Op("endure_fn_providers_call")
	ret := f.Func.Call(in)
	for i := 0; i < len(ret); i++ {
		// try to find possible errors
		r := ret[i].Interface()
		if r == nil {
			continue
		}
		if rErr, ok := r.(error); ok {
			if rErr != nil {
				e.logger.Error("error occurred in the traverseCallProvider", zap.String("id", vertex.ID))
				return errors.E(op, errors.FunctionCall, rErr)
			}
			continue
		}

		// add the value to the Providers
		e.logger.Debug("value added successfully", zap.String("id", vertex.ID), zap.String("caller id", callerID), zap.String("parameter", ret[i].Type().String()))
		e.graph.AddGlobalProvider(removePointerAsterisk(ret[i].Type().String()), ret[i])
		vertex.AddProvider(removePointerAsterisk(ret[i].Type().String()), ret[i], isReference(ret[i].Type()), ret[i].Kind())
	}

	return nil
}
