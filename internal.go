package endure

import (
	"context"
	"fmt"
	"reflect"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/spiral/endure/errors"
	"github.com/spiral/endure/structures"
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

	if len(vertex.Meta.DependsDepsToInvoke) > 0 {
		for i := 0; i < len(vertex.Meta.DependsDepsToInvoke); i++ {
			// Interface dependency
			if vertex.Meta.DependsDepsToInvoke[i].Kind == reflect.Interface {
				err = e.traverseCallDependersInterface(vertex)
				if err != nil {
					return errors.E(op, errors.Traverse, err)
				}
			} else {
				// structure dependence
				err = e.traverseCallDependers(vertex)
				if err != nil {
					return errors.E(op, errors.Traverse, err)
				}
			}
		}
	}
	return nil
}

func (e *Endure) traverseCallDependersInterface(vertex *structures.Vertex) error {
	const op = errors.Op("internal_traverse_call_dependers_interface")
	for i := 0; i < len(vertex.Meta.DependsDepsToInvoke); i++ {
		// get dependency id (vertex id)
		depID := vertex.Meta.DependsDepsToInvoke[i].Name
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
				// init
				inInterface := make([]reflect.Value, 0, 2)
				// add service itself
				inInterface = append(inInterface, reflect.ValueOf(vertex.Iface))
				// if type provides needed type
				// value - reference and init dep also reference
				switch {
				case *vertexVal.IsReference == *vertex.Meta.DependsDepsToInvoke[i].IsReference:
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

				err := e.callDependerFns(vertex, inInterface)
				if err != nil {
					return errors.E(op, errors.Traverse, err)
				}
			}
		}
	}

	return nil
}

func (e *Endure) traverseCallDependers(vertex *structures.Vertex) error {
	const op = "internal_traverse_call_dependers"
	in := make([]reflect.Value, 0, 2)
	// add service itself
	in = append(in, reflect.ValueOf(vertex.Iface))

	for i := 0; i < len(vertex.Meta.DependsDepsToInvoke); i++ {
		// get dependency id (vertex id)
		depID := vertex.Meta.DependsDepsToInvoke[i].Name
		// find vertex which provides dependency
		providers := e.graph.FindProviders(depID)
		// search for providers
		for j := 0; j < len(providers); j++ {
			for vertexID, val := range providers[j].Provides {
				// if type provides needed type
				if vertexID == depID {
					switch {
					case *val.IsReference == *vertex.Meta.DependsDepsToInvoke[i].IsReference:
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

	err := e.callDependerFns(vertex, in)
	if err != nil {
		return errors.E(op, errors.Traverse, err)
	}

	return nil
}

func (e *Endure) callDependerFns(vertex *structures.Vertex, in []reflect.Value) error {
	const op = errors.Op("internal_call_depender_functions")
	// type implements Depender interface
	if reflect.TypeOf(vertex.Iface).Implements(reflect.TypeOf((*Depender)(nil)).Elem()) {
		// if type implements Depender() it should has FnsProviderToInvoke
		if vertex.Meta.DependsDepsToInvoke != nil {
			for k := 0; k < len(vertex.Meta.FnsDependerToInvoke); k++ {
				m, ok := reflect.TypeOf(vertex.Iface).MethodByName(vertex.Meta.FnsDependerToInvoke[k])
				if !ok {
					e.logger.Error("type has missing method in FnsDependerToInvoke", zap.String("vertex id", vertex.ID), zap.String("method", vertex.Meta.FnsDependerToInvoke[k]))
					return errors.E(op, errors.FunctionCall, errors.Str("type has missing method in FnsDependerToInvoke"))
				}

				ret := m.Func.Call(in)
				// handle error
				if len(ret) > 0 {
					// error is the last return parameter in line
					rErr := ret[len(ret)-1].Interface()
					if rErr != nil {
						if err, ok := rErr.(error); ok && e != nil {
							e.logger.Error("error calling DependerFns", zap.String("vertex id", vertex.ID), zap.Error(err))
							return errors.E(op, errors.FunctionCall, err)
						}
						return errors.E(op, errors.FunctionCall, errors.Str("unknown error occurred during the function call"))
					}
				} else {
					return errors.E(op, errors.FunctionCall, errors.Str("depender should return Value and error types"))
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

func (e *Endure) traverseProviders(depsEntry structures.Entry, depVertex *structures.Vertex, depID string, calleeID string, in []reflect.Value) ([]reflect.Value, error) {
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

func (e *Endure) appendProviderFuncArgs(depsEntry structures.Entry, providedEntry structures.ProvidedEntry, in []reflect.Value) []reflect.Value {
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

func (e *Endure) traverseCallProvider(vertex *structures.Vertex, in []reflect.Value, callerID, depId string) error {
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

/*
Algorithm is the following (all steps executing in the topological order):
2. Call Serve() on all services --     OPTIONAL
3. Call Stop() on all services --      OPTIONAL
4. Call Clear() on a services, which implements this interface -- OPTIONAL
*/
// call configure on the node

func (e *Endure) callServeFn(vertex *structures.Vertex, in []reflect.Value) *result {
	m, _ := reflect.TypeOf(vertex.Iface).MethodByName(ServeMethodName)
	ret := m.Func.Call(in)
	res := ret[0].Interface()
	if res != nil {
		e.logger.Debug("start serving vertex", zap.String("vertexId", vertex.ID))
		if e, ok := res.(chan error); ok && e != nil {
			return &result{
				errCh:    e,
				signal:   make(chan notify),
				vertexID: vertex.ID,
			}
		}
	}
	// error, result should not be nil
	// the only one reason to be nil is to vertex return parameter (channel) is not initialized
	return nil
}

func (e *Endure) stop(vID string) error {
	const op = errors.Op("internal_stop")
	vertex := e.graph.GetVertex(vID)
	if reflect.TypeOf(vertex.Iface).Implements(reflect.TypeOf((*Service)(nil)).Elem()) {
		in := make([]reflect.Value, 0, 1)
		// add service itself
		in = append(in, reflect.ValueOf(vertex.Iface))

		err := e.callStopFn(vertex, in)
		if err != nil {
			e.logger.Error("error occurred during the callStopFn", zap.String("vertex id", vertex.ID))
			return errors.E(op, errors.FunctionCall, err)
		}
	}

	return nil
}

func (e *Endure) callStopFn(vertex *structures.Vertex, in []reflect.Value) error {
	const op = errors.Op("internal_call_stop_function")
	// Call Stop() method, which returns only error (or nil)
	e.logger.Debug("stopping vertex", zap.String("vertexId", vertex.ID))
	m, _ := reflect.TypeOf(vertex.Iface).MethodByName(StopMethodName)
	ret := m.Func.Call(in)
	rErr := ret[0].Interface()
	if rErr != nil {
		if e, ok := rErr.(error); ok && e != nil {
			return errors.E(op, errors.FunctionCall, e)
		}
		return errors.E(op, errors.FunctionCall, errors.Str("unknown error occurred during the function call"))
	}
	return nil
}

func (e *Endure) sendStopSignal(sorted []*structures.Vertex) {
	for _, v := range sorted {
		// get result by vertex ID
		tmp, ok := e.results.Load(v.ID)
		if !ok {
			continue
		}
		res := tmp.(*result)
		if tmp == nil {
			continue
		}
		// send exit signal to the goroutine in sorted order
		e.logger.Debug("sending exit signal to the vertex from the main thread", zap.String("vertex id", res.vertexID))
		res.signal <- notify{
			stop: true,
		}

		e.results.Delete(v.ID)
	}
}

func (e *Endure) sendResultToUser(res *result) {
	e.userResultsCh <- &Result{
		Error:    res.err,
		VertexID: res.vertexID,
	}
}

func (e *Endure) shutdown(n *structures.DllNode) {
	// channel with nodes to stop
	sh := make(chan *structures.DllNode)
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute*5)
	go func() {
		// process all nodes one by one
		nCopy := n
		for nCopy != nil {
			if nCopy.Vertex.IsDisabled == true {
				nCopy = nCopy.Next
				continue
			}
			sh <- nCopy
			nCopy = nCopy.Next
		}
		// after all nodes will be processed, send ctx.Done signal to finish the stopHandler
		cancel()
	}()
	// block until process all nodes
	e.forceExitHandler(ctx, sh)
}

func (e *Endure) forceExitHandler(ctx context.Context, data chan *structures.DllNode) {
	for {
		select {
		case node := <-data:
			// stop vertex
			err := e.stop(node.Vertex.ID)
			if err != nil {
				// TODO do not return until finished
				// just log the errors
				// stack it in slice and if slice is not empty, print it ??
				e.logger.Error("error occurred during the services stopping", zap.String("vertex id", node.Vertex.ID), zap.Error(err))
			}
			// exit from vertex poller
			tmp, ok := e.results.Load(node.Vertex.ID)
			if !ok {
				continue
			}

			channel := tmp.(*result)
			channel.signal <- notify{
				// false because we called stop already
				stop: false,
			}

		case <-ctx.Done():
			return
		}
	}
}

// serve run calls callServeFn for each node and put the results in the map
func (e *Endure) serve(n *structures.DllNode) error {
	const op = errors.Op("internal_serve")
	// check if type implements serve, if implements, call serve
	if reflect.TypeOf(n.Vertex.Iface).Implements(reflect.TypeOf((*Service)(nil)).Elem()) {
		in := make([]reflect.Value, 0, 1)
		// add service itself
		in = append(in, reflect.ValueOf(n.Vertex.Iface))

		res := e.callServeFn(n.Vertex, in)
		if res != nil {
			e.results.Store(res.vertexID, res)
		} else {
			e.logger.Error("nil result returned from the vertex", zap.String("vertex id", n.Vertex.ID), zap.String("tip:", "serve function should return initialized channel with errors"))
			return errors.E(op, errors.FunctionCall, errors.Errorf("nil result returned from the vertex, vertex id: %s", n.Vertex.ID))
		}

		// start poll the vertex
		e.poll(res)
	}

	return nil
}

func (e *Endure) startMainThread() {
	/*
		Main thread is the main Endure unit of work
		It used to handle errors from vertices, notify user about result, re-calculating graph according to failed vertices and sending stop signals
	*/
	go func() {
		for {
			select {
			// failed Vertex
			case res, ok := <-e.handleErrorCh:
				// lock the handleErrorCh processing
				if !ok {
					e.logger.Debug("handle error channel was closed")
					return
				}

				e.logger.Debug("processing error in the main thread", zap.String("vertex id", res.vertexID))
				if e.retry {
					// TODO handle error from the retry handler
					e.retryHandler(res)
				} else {
					e.logger.Info("retry is turned off, sending exit signal to every vertex in the graph")
					// send exit signal to whole graph
					e.sendStopSignal(e.graph.Vertices)
					e.sendResultToUser(res)
				}
			}
		}
	}()
}

func (e *Endure) retryHandler(res *result) {
	const op = errors.Op("internal_retry_handler")
	// get vertex from the graph
	vertex := e.graph.GetVertex(res.vertexID)
	if vertex == nil {
		e.logger.Error("failed to get vertex from the graph, vertex is nil", zap.String("vertex id from the handleErrorCh channel", res.vertexID))
		e.userResultsCh <- &Result{
			Error:    errors.E(op, errors.Traverse, errors.Str("failed to get vertex from the graph, vertex is nil")),
			VertexID: "",
		}
		return
	}

	// reset vertex and dependencies to the initial state
	// numOfDeps and visited/visiting
	vertices := e.graph.Reset(vertex)

	// Topologically sort the graph
	sorted, err := structures.TopologicalSort(vertices)
	if err != nil {
		e.logger.Error("error sorting the graph", zap.Error(err))
		return
	}
	if sorted == nil {
		e.logger.Error("sorted list should not be nil", zap.String("vertex id from the handleErrorCh channel", res.vertexID))
		e.userResultsCh <- &Result{
			Error:    errors.E(op, errors.Traverse, errors.Str("failed to topologically sort the graph")),
			VertexID: res.vertexID,
		}
		return
	}

	// send exit signal only to sorted and involved vertices
	// stop will be called inside poller
	e.sendStopSignal(sorted)

	// Init backoff
	b := backoff.NewExponentialBackOff()
	b.MaxElapsedTime = e.maxInterval
	b.InitialInterval = e.initialInterval

	affectedRunList := structures.NewDoublyLinkedList()
	for i := 0; i <= len(sorted)-1; i++ {
		affectedRunList.Push(sorted[i])
	}

	// call init
	headCopy := affectedRunList.Head
	for headCopy != nil {
		berr := backoff.Retry(e.backoffInit(headCopy.Vertex), b)
		if berr != nil {
			e.logger.Error("backoff failed", zap.String("vertex id", headCopy.Vertex.ID), zap.Error(berr))
			e.userResultsCh <- &Result{
				Error:    errors.E(op, errors.FunctionCall, errors.Errorf("error during the Init function call")),
				VertexID: headCopy.Vertex.ID,
			}
			return
		}
		headCopy = headCopy.Next
	}

	// call serve
	headCopy = affectedRunList.Head
	for headCopy != nil {
		err := e.serve(headCopy)
		if err != nil {
			e.userResultsCh <- &Result{
				Error:    errors.E(op, errors.FunctionCall, errors.Errorf("error during the Serve function call")),
				VertexID: headCopy.Vertex.ID,
			}
			e.logger.Error("fatal error during the serve in the main thread", zap.String("vertex id", headCopy.Vertex.ID), zap.Error(err))
			return
		}

		headCopy = headCopy.Next
	}

	e.sendResultToUser(res)
}

// poll is used to poll the errors from the vertex
// and exit from it
func (e *Endure) poll(r *result) {
	rr := r
	go func(res *result) {
		for {
			select {
			// error
			case err := <-res.errCh:
				if err != nil {
					// log error message
					e.logger.Error("vertex got an error", zap.String("vertex id", res.vertexID), zap.Error(err))

					// set the error
					res.err = err

					// send handleErrorCh signal
					e.handleErrorCh <- res
				}
			// exit from the goroutine
			case n := <-res.signal:
				if n.stop {
					e.mutex.Lock()
					e.logger.Info("vertex got exit signal", zap.String("vertex id", res.vertexID))
					err := e.stop(res.vertexID)
					if err != nil {
						e.logger.Error("error during exit signal", zap.String("error while stopping the vertex:", res.vertexID), zap.Error(err))
						e.mutex.Unlock()
					}
					e.mutex.Unlock()
					return
				}
				return
			}
		}
	}(rr)
}

func (e *Endure) register(name string, vertex interface{}, order int) error {
	// check the vertex
	const op = errors.Op("internal_register")
	if e.graph.HasVertex(name) {
		return errors.E(op, errors.Traverse, errors.Errorf("vertex `%s` already exists", name))
	}

	meta := structures.Meta{
		Order: order,
	}

	// just push the vertex
	// here we can append in future some meta information
	e.graph.AddVertex(name, vertex, meta)
	return nil
}

func (e *Endure) backoffInit(v *structures.Vertex) func() error {
	return func() error {
		const op = errors.Op("internal_backoff_init")
		// we already checked the Interface satisfaction
		// at this step absence of Init() is impossible
		init, _ := reflect.TypeOf(v.Iface).MethodByName(InitMethodName)

		err := e.callInitFn(init, v)
		if err != nil {
			e.logger.Error("error occurred during the call INIT function", zap.String("vertex id", v.ID), zap.Error(err))
			return errors.E(op, errors.FunctionCall, err)
		}

		return nil
	}
}
