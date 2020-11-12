package endure

import (
	"fmt"
	"reflect"
	"sort"

	"github.com/spiral/errors"
	"go.uber.org/zap"
)

func (e *Endure) traverseProviders(depsEntry Entry, fnReceiver *Vertex, depID string, calleeVertexId string, in []reflect.Value) ([]reflect.Value, error) {
	const op = errors.Op("internal_traverse_providers")
	err := e.traverseCallProvider(fnReceiver, []reflect.Value{reflect.ValueOf(fnReceiver.Iface)}, calleeVertexId, depID)
	if err != nil {
		return nil, errors.E(op, errors.Traverse, err)
	}

	// to index function fn in defer
	for providerID, providedEntry := range fnReceiver.Provides {
		if providerID == depID {
			in = e.appendProviderFuncArgs(depsEntry, providedEntry, in)
		}
	}

	return in, nil
}

func (e *Endure) appendProviderFuncArgs(depsEntry Entry, providedEntry ProvidedEntry, in []reflect.Value) []reflect.Value {
	switch {
	case *providedEntry.IsReference == *depsEntry.IsReference:
		in = append(in, providedEntry.Value)
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

type Providers []Provide

//
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

func (e *Endure) traverseCallProvider(fnReceiver *Vertex, in []reflect.Value, callerID, depId string) error {
	const op = errors.Op("internal_traverse_call_provider")
	providers := make(Providers, 0, 0)
	// to index function fn in defer
	i := 0
	defer func() {
		if r := recover(); r != nil {
			e.logger.Error("panic during the function call", zap.String("function fn", fnReceiver.Meta.FnsProviderToInvoke[i].FunctionName), zap.String("error", fmt.Sprint(r)))
		}
	}()

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
			for i := 0; i < len(fnReceiver.Meta.FnsProviderToInvoke); i++ {
				p := Provide{}
				m, ok := reflect.TypeOf(fnReceiver.Iface).MethodByName(fnReceiver.Meta.FnsProviderToInvoke[i].FunctionName)
				if !ok {
					e.logger.Panic("should implement the Provider interface", zap.String("function fn", fnReceiver.Meta.FnsProviderToInvoke[i].FunctionName))
				}
				p.m = m

				if fnReceiver.Meta.FnsProviderToInvoke[i].ReturnTypeId != depId {
					continue
				}

				// example ProvideWithName(named endure.Named) (SuperInterface, error)
				// IN -> endure.Named + struct receiver
				// OUT -> SuperInterface, error
				for f := 0; f < m.Func.Type().NumIn(); f++ {
					p.In = append(p.In, m.Func.Type().In(f))
				}

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
			for j := 0; j < len(providers); j++ {
				pr := providers[j]

				// fallback call provided, only 1 IN arg, function receiver
				if len(pr.In) == 1 {
					return e.fnCall(pr.m, in, fnReceiver, callerID)
				}

				// if we have minimum 2 In args (self and Named for example)
				// we should check where is function receiver and check if caller implement all other args
				// if everything ok we just pass first args as the receiver and caller as all the rest args
				// start from 1, 0-th index is function receiver
				// check if caller implements all needed interfaces to call func
				if e.walk(pr.In, callerV) == false {
					// if not, check for other provider
					continue
				}

				for k := 1; k < len(pr.In); k++ {
					switch pr.In[k].Kind() {
					case reflect.Struct: // just structure
						in = append(in, e.graph.providers[pr.In[k].String()])
					case reflect.Ptr: // Ptr to the structure
						val := pr.In[k].Elem() // get real value
						in = append(in, e.graph.providers[val.String()])
					case reflect.Interface: // we know here, that caller implement all needed to call interfaces
						in = append(in, reflect.ValueOf(e.graph.VerticesMap[callerID].Iface))
					}
				}

				return e.fnCall(pr.m, in, fnReceiver, callerID)
			}

			return errors.E("can't call provides, no suitable vertex")
		}
	}
	return nil
}

func (e *Endure) fnCall(f reflect.Method, in []reflect.Value, vertex *Vertex, callerId string) error {
	const op = errors.Op("provider fn call")
	ret := f.Func.Call(in)
	// handle error
	//if len(ret) > 1 {
	for i := 0; i < len(ret); i++ {
		// try to find possible errors
		r := ret[i].Interface()
		if r == nil {
			continue
		}
		if rErr, ok := r.(error); ok {
			if rErr != nil {
				if err, ok := rErr.(error); ok && e != nil {
					e.logger.Error("error occurred in the traverseCallProvider", zap.String("vertex id", vertex.ID))
					return errors.E(op, errors.FunctionCall, err)
				}
				return errors.E(op, errors.FunctionCall, errors.Str("unknown error occurred during the function call"))
			}
			continue
		}

		// add the value to the Providers
		e.logger.Debug("value added successfully", zap.String("vertex id", vertex.ID), zap.String("caller id", callerId), zap.String("parameter", ret[i].Type().String()))
		e.graph.AddGlobalProvider(removePointerAsterisk(ret[i].Type().String()), ret[i])
		vertex.AddProvider(removePointerAsterisk(ret[i].Type().String()), ret[i], isReference(ret[i].Type()), ret[i].Kind())
	}

	return nil
}
