package endure

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/spiral/endure/structures"
	"go.uber.org/zap"
)

/*
   Traverse the DLL in the forward direction

*/
func (e *Endure) init(vertex *structures.Vertex) error {
	// we already checked the Interface satisfaction
	// at this step absence of Init() is impoosssibruuu
	initMethod, _ := reflect.TypeOf(vertex.Iface).MethodByName(InitMethodName)

	err := e.callInitFn(initMethod, vertex)
	if err != nil {
		e.logger.Error("error occurred during the call INIT function", zap.String("vertex id", vertex.ID), zap.Error(err))
		return err
	}

	return nil
}

func (e *Endure) callInitFn(init reflect.Method, vertex *structures.Vertex) error {
	defer func() {
		if r := recover(); r != nil {
			e.logger.Error("[panic][recovered] probably called Init with insufficient number of params. check the init function and make sure you are registered dependency")
			// continue panic to prevent user to use Serve
			panic("probably called Init with insufficient number of params. check the init function and make sure you are registered dependency")
		}
	}()
	in, err := e.findInitParameters(vertex)
	if err != nil {
		return err
	}
	// Iterate over dependencies
	// And search in Vertices for the provided types
	ret := init.Func.Call(in)
	rErr := ret[0].Interface()
	if rErr != nil {
		if err, ok := rErr.(error); ok && e != nil {
			e.logger.Error("error calling init", zap.String("vertex id", vertex.ID), zap.Error(err))
			return err
		}
		return errUnknownErrorOccurred
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
		return errors.New("0 or less parameters for Init")
	}

	if len(vertex.Meta.DepsList) > 0 {
		for i := 0; i < len(vertex.Meta.DepsList); i++ {
			// Interface dependency
			if vertex.Meta.DepsList[i].Kind == reflect.Interface {
				err = e.traverseCallDependersInterface(vertex)
				if err != nil {
					return err
				}
			} else {
				// structure dependence
				err = e.traverseCallDependers(vertex)
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func (e *Endure) traverseCallDependersInterface(vertex *structures.Vertex) error {
	for i := 0; i < len(vertex.Meta.DepsList); i++ {
		// get dependency id (vertex id)
		depID := vertex.Meta.DepsList[i].Name
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
				case *vertexVal.IsReference == *vertex.Meta.DepsList[i].IsReference:
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
					return err
				}
			}
		}
	}

	return nil
}

func (e *Endure) traverseCallDependers(vertex *structures.Vertex) error {
	in := make([]reflect.Value, 0, 2)
	// add service itself
	in = append(in, reflect.ValueOf(vertex.Iface))

	for i := 0; i < len(vertex.Meta.DepsList); i++ {
		// get dependency id (vertex id)
		depID := vertex.Meta.DepsList[i].Name
		// find vertex which provides dependency
		providers := e.graph.FindProviders(depID)
		// search for providers
		for j := 0; j < len(providers); j++ {
			for vertexID, val := range providers[j].Provides {
				// if type provides needed type
				if vertexID == depID {
					switch {
					case *val.IsReference == *vertex.Meta.DepsList[i].IsReference:
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
		return err
	}

	return nil
}

func (e *Endure) callDependerFns(vertex *structures.Vertex, in []reflect.Value) error {
	// type implements Depender interface
	if reflect.TypeOf(vertex.Iface).Implements(reflect.TypeOf((*Depender)(nil)).Elem()) {
		// if type implements Depender() it should has FnsProviderToInvoke
		if vertex.Meta.DepsList != nil {
			for k := 0; k < len(vertex.Meta.FnsDependerToInvoke); k++ {
				m, ok := reflect.TypeOf(vertex.Iface).MethodByName(vertex.Meta.FnsDependerToInvoke[k])
				if !ok {
					e.logger.Error("type has missing method in FnsDependerToInvoke", zap.String("vertex id", vertex.ID), zap.String("method", vertex.Meta.FnsDependerToInvoke[k]))
					return errors.New("type has missing method in FnsDependerToInvoke")
				}

				ret := m.Func.Call(in)
				// handle error
				if len(ret) > 0 {
					// error is the last return parameter in line
					rErr := ret[len(ret)-1].Interface()
					if rErr != nil {
						if err, ok := rErr.(error); ok && e != nil {
							e.logger.Error("error calling DependerFns", zap.String("vertex id", vertex.ID), zap.Error(err))
							return err
						}
						return errUnknownErrorOccurred
					}
				} else {
					return errors.New("depender should return Value and error types")
				}
			}
		}
	}
	return nil
}

func (e *Endure) findInitParameters(vertex *structures.Vertex) ([]reflect.Value, error) {
	in := make([]reflect.Value, 0, 2)

	// add service itself
	in = append(in, reflect.ValueOf(vertex.Iface))

	// add dependencies
	if len(vertex.Meta.InitDepsList) > 0 {
		for i := 0; i < len(vertex.Meta.InitDepsList); i++ {
			depID := vertex.Meta.InitDepsList[i].Name
			v := e.graph.FindProviders(depID)
			var err error
			in, err = e.traverseProviders(vertex.Meta.InitDepsList[i], v[0], depID, vertex.ID, in)
			if err != nil {
				return nil, err
			}
		}
	}
	return in, nil
}

func (e *Endure) traverseProviders(depsEntry structures.DepsEntry, depVertex *structures.Vertex, depID string, calleeID string, in []reflect.Value) ([]reflect.Value, error) {
	// we need to call all providers first
	err := e.traverseCallProvider(depVertex, []reflect.Value{reflect.ValueOf(depVertex.Iface)}, calleeID)
	if err != nil {
		return nil, err
	}

	// to index function name in defer
	for providerID, providedEntry := range depVertex.Provides {
		if providerID == depID {
			in = e.appendProviderFuncArgs(depsEntry, providedEntry, in)
		}
	}

	return in, nil
}

func (e *Endure) appendProviderFuncArgs(depsEntry structures.DepsEntry, providedEntry structures.ProvidedEntry, in []reflect.Value) []reflect.Value {
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

func (e *Endure) traverseCallProvider(vertex *structures.Vertex, in []reflect.Value, callerID string) error {
	// to index function name in defer
	i := 0
	defer func() {
		if r := recover(); r != nil {
			e.logger.Error("panic during the function call", zap.String("function name", vertex.Meta.FnsProviderToInvoke[i]), zap.String("error", fmt.Sprint(r)))
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
				m, ok := reflect.TypeOf(vertex.Iface).MethodByName(vertex.Meta.FnsProviderToInvoke[i])
				if !ok {
					e.logger.Panic("should implement the Provider interface", zap.String("function name", vertex.Meta.FnsProviderToInvoke[i]))
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
					at the moment we assume, that this "other type" is Name interface
				*/
				if m.Func.Type().NumIn() > 1 {
					/*
						here we should add type which implement Named interface
						at the moment we seek for implementation in the callerID only
					*/

					callerV := e.graph.GetVertex(callerID)
					if callerV == nil {
						return errors.New("caller vertex is nil")
					}

					// skip function receiver
					for j := 1; j < m.Func.Type().NumIn(); j++ {
						// current function IN type (interface)
						t := m.Func.Type().In(j)
						if t.Kind() != reflect.Interface {
							e.logger.Panic("Provider accepts only interfaces", zap.String("function name", vertex.Meta.FnsProviderToInvoke[i]))
						}

						// if Caller struct implements interface -- ok, add it to the inCopy list
						// else panic
						if reflect.TypeOf(callerV.Iface).Implements(t) == false {
							e.logger.Panic("Caller should implement callee interface", zap.String("function name", vertex.Meta.FnsProviderToInvoke[i]))
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
							return err
						}
						return errUnknownErrorOccurred
					}

					// add the value to the Providers
					e.logger.Debug("value added successfully", zap.String("vertex id", vertex.ID), zap.String("caller id", callerID), zap.String("parameter", in[0].Type().String()))
					vertex.AddProvider(removePointerAsterisk(ret[0].Type().String()), ret[0], isReference(ret[0].Type()), in[0].Kind())
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

/*
callConfigureFn invoke Configure() error method
*/
func (e *Endure) callConfigureFn(vertex *structures.Vertex, in []reflect.Value) error {
	m, _ := reflect.TypeOf(vertex.Iface).MethodByName(ConfigureMethodName)
	ret := m.Func.Call(in)
	res := ret[0].Interface()
	if res != nil {
		if e, ok := res.(error); ok && e != nil {
			return e
		}
		return errUnknownErrorOccurred
	}
	return nil
}

func (e *Endure) callStopFn(vertex *structures.Vertex, in []reflect.Value) error {
	// Call Stop() method, which returns only error (or nil)
	e.logger.Debug("stopping vertex", zap.String("vertexId", vertex.ID))
	m, _ := reflect.TypeOf(vertex.Iface).MethodByName(StopMethodName)
	ret := m.Func.Call(in)
	rErr := ret[0].Interface()
	if rErr != nil {
		if e, ok := rErr.(error); ok && e != nil {
			return e
		}
		return errUnknownErrorOccurred
	}
	return nil
}

func (e *Endure) stop(vID string) error {
	vertex := e.graph.GetVertex(vID)

	in := make([]reflect.Value, 0, 1)
	// add service itself
	in = append(in, reflect.ValueOf(vertex.Iface))

	err := e.callStopFn(vertex, in)
	if err != nil {
		e.logger.Error("error occurred during the callStopFn", zap.String("vertex id", vertex.ID))
		return err
	}

	if reflect.TypeOf(vertex.Iface).Implements(reflect.TypeOf((*Graceful)(nil)).Elem()) {
		err = e.callCloseFn(vertex.ID, in)
		if err != nil {
			e.logger.Error("error occurred during the callCloseFn", zap.String("vertex id", vertex.ID))
			return err
		}
	}

	return nil
}

// TODO add stack to the all of the log events
func (e *Endure) callCloseFn(vID string, in []reflect.Value) error {
	v := e.graph.GetVertex(vID)
	// Call Close() method, which returns only error (or nil)
	m, _ := reflect.TypeOf(v.Iface).MethodByName(CloseMethodName)
	ret := m.Func.Call(in)
	rErr := ret[0].Interface()
	if rErr != nil {
		if e, ok := rErr.(error); ok && e != nil {
			return e
		}
		return errUnknownErrorOccurred
	}
	return nil
}

func (e *Endure) sendStopSignal(sorted []*structures.Vertex) {
	for _, v := range sorted {
		// get result by vertex ID
		tmp := e.results[v.ID]
		if tmp == nil {
			continue
		}
		// send exit signal to the goroutine in sorted order
		e.logger.Debug("sending exit signal to the vertex from the main thread", zap.String("vertex id", tmp.vertexID))
		tmp.signal <- notify{
			stop: true,
		}

		e.results[v.ID] = nil
	}
}

func (e *Endure) sendResultToUser(res *result) {
	e.userResultsCh <- &Result{
		Error: Error{
			Err:   res.err,
			Code:  0,
			Stack: nil,
		},
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
			if channel, ok := e.results[node.Vertex.ID]; (ok == true) && (channel != nil) {
				channel.signal <- notify{
					// false because we called stop already
					stop: false,
				}
			}

		case <-ctx.Done():
			return
		}
	}
}

// serve run configure (if exist) and callServeFn for each node and put the results in the map
func (e *Endure) serve(n *structures.DllNode) error {
	// handle all configure
	in := make([]reflect.Value, 0, 1)
	// add service itself
	in = append(in, reflect.ValueOf(n.Vertex.Iface))

	res := e.callServeFn(n.Vertex, in)
	if res != nil {
		e.results[res.vertexID] = res
	} else {
		e.logger.Error("nil result returned from the vertex", zap.String("vertex id", n.Vertex.ID), zap.String("tip:", "serve function should return initialized channel with errors"))
		return fmt.Errorf("nil result returned from the vertex, vertex id: %s", n.Vertex.ID)
	}

	e.poll(res)
	if e.restartedTime[n.Vertex.ID] != nil {
		*e.restartedTime[n.Vertex.ID] = time.Now()
	} else {
		tmp := time.Now()
		e.restartedTime[n.Vertex.ID] = &tmp
	}

	return nil
}

func (e *Endure) checkLeafErrorTime(res *result) bool {
	e.mutex.Lock()
	defer e.mutex.Unlock()
	return e.restartedTime[res.vertexID] != nil && e.restartedTime[res.vertexID].After(*e.errorTime[res.vertexID])
}

func (e *Endure) startMainThread() {
	/*
		Used for handling error from the vertices
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
				if e.checkLeafErrorTime(res) {
					e.logger.Debug("error processing skipped because vertex already restarted by the root", zap.String("vertex id", res.vertexID))
					e.sendResultToUser(res)
					continue
				}

				// get vertex from the graph
				vertex := e.graph.GetVertex(res.vertexID)
				if vertex == nil {
					e.logger.Error("failed to get vertex from the graph, vertex is nil", zap.String("vertex id from the handleErrorCh channel", res.vertexID))
					e.userResultsCh <- &Result{
						Error:    FailedToGetTheVertex,
						VertexID: "",
					}
					return
				}

				// reset vertex and dependencies to the initial state
				// numOfDeps and visited/visiting
				vertices := e.graph.Reset(vertex)

				// Topologically sort the graph
				sorted := structures.TopologicalSort(vertices)
				if sorted == nil {
					e.logger.Error("sorted list should not be nil", zap.String("vertex id from the handleErrorCh channel", res.vertexID))
					e.userResultsCh <- &Result{
						Error:    FailedToSortTheGraph,
						VertexID: res.vertexID,
					}
					return
				}

				if e.retry {
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
								Error:    ErrorDuringInit,
								VertexID: headCopy.Vertex.ID,
							}
							return
						}

						headCopy = headCopy.Next
					}

					// call configure
					headCopy = affectedRunList.Head
					for headCopy != nil {
						berr := backoff.Retry(e.backoffConfigure(headCopy), b)
						if berr != nil {
							e.userResultsCh <- &Result{
								Error:    ErrorDuringInit,
								VertexID: headCopy.Vertex.ID,
							}
							e.logger.Error("backoff failed", zap.String("vertex id", headCopy.Vertex.ID), zap.Error(berr))
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
								Error:    ErrorDuringServe,
								VertexID: headCopy.Vertex.ID,
							}
							e.logger.Error("fatal error during the serve in the main thread", zap.String("vertex id", headCopy.Vertex.ID), zap.Error(err))
							return
						}

						headCopy = headCopy.Next
					}

					e.sendResultToUser(res)
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
					// set error time
					e.mutex.Lock()
					if e.errorTime[res.vertexID] != nil {
						*e.errorTime[res.vertexID] = time.Now()
					} else {
						tmp := time.Now()
						e.errorTime[res.vertexID] = &tmp
					}
					e.mutex.Unlock()

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
	if e.graph.HasVertex(name) {
		return errVertexAlreadyExists(name)
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
		// we already checked the Interface satisfaction
		// at this step absence of Init() is impossible
		init, _ := reflect.TypeOf(v.Iface).MethodByName(InitMethodName)

		err := e.callInitFn(init, v)
		if err != nil {
			e.logger.Error("error occurred during the call INIT function", zap.String("vertex id", v.ID), zap.Error(err))
			return err
		}

		return nil
	}
}

func (e *Endure) configure(n *structures.DllNode) error {
	// handle all configure
	in := make([]reflect.Value, 0, 1)
	// add service itself
	in = append(in, reflect.ValueOf(n.Vertex.Iface))

	if reflect.TypeOf(n.Vertex.Iface).Implements(reflect.TypeOf((*Graceful)(nil)).Elem()) {
		err := e.callConfigureFn(n.Vertex, in)
		if err != nil {
			return err
		}
	}

	return nil
}

func (e *Endure) backoffConfigure(n *structures.DllNode) func() error {
	return func() error {
		// handle all configure
		in := make([]reflect.Value, 0, 1)
		// add service itself
		in = append(in, reflect.ValueOf(n.Vertex.Iface))

		if reflect.TypeOf(n.Vertex.Iface).Implements(reflect.TypeOf((*Graceful)(nil)).Elem()) {
			err := e.callConfigureFn(n.Vertex, in)
			if err != nil {
				e.logger.Error("error configuring the vertex", zap.String("vertex id", n.Vertex.ID), zap.Error(err))
				return err
			}
		}

		return nil
	}
}
