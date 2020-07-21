package cascade

import (
	"errors"
	"fmt"
	"reflect"
	"time"

	"github.com/spiral/cascade/structures"
	"go.uber.org/zap"
)

/*
   Traverse the DLL in the forward direction

*/
func (c *Cascade) init(vertex *structures.Vertex) error {
	// we already checked the Interface satisfaction
	// at this step absence of Init() is impoosssibruuu
	initMethod, _ := reflect.TypeOf(vertex.Iface).MethodByName(InitMethodName)

	err := c.callInitFn(initMethod, vertex)
	if err != nil {
		c.logger.Error("error occurred during the call INIT function", zap.String("vertex id", vertex.ID), zap.Error(err))
		return err
	}

	return nil
}

func (c *Cascade) callInitFn(init reflect.Method, vertex *structures.Vertex) error {
	in, err := c.findInitParameters(vertex)
	if err != nil {
		return err
	}
	// Iterate over dependencies
	// And search in Vertices for the provided types
	ret := init.Func.Call(in)
	rErr := ret[0].Interface()
	if rErr != nil {
		if e, ok := rErr.(error); ok && e != nil {
			c.logger.Error("error calling init", zap.String("vertex id", vertex.ID), zap.Error(e))
			return e
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
		err := vertex.AddProvider(removePointerAsterisk(in[0].Type().String()), in[0], isReference(in[0].Type()), in[0].Kind())
		if err != nil {
			c.logger.Debug("value added successfully", zap.String("vertex id", vertex.ID), zap.String("parameter", in[0].Type().String()))
			return err
		}
	} else {
		c.logger.Error("0 or less parameters for Init", zap.String("vertex id", vertex.ID))
		return errors.New("0 or less parameters for Init")
	}

	if len(vertex.Meta.DepsList) > 0 {
		for i := 0; i < len(vertex.Meta.DepsList); i++ {
			// Interface dependency
			if vertex.Meta.DepsList[i].Kind == reflect.Interface {
				err := c.traverseCallDependersInterface(vertex)
				if err != nil {
					return err
				}
			} else {
				// structure dependence
				err := c.traverseCallDependers(vertex)
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func (c *Cascade) traverseCallDependersInterface(vertex *structures.Vertex) error {
	for i := 0; i < len(vertex.Meta.DepsList); i++ {
		// get dependency id (vertex id)
		depID := vertex.Meta.DepsList[i].Name
		// find vertex which provides dependency
		providers := c.graph.FindProviders(depID)

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
						c.logger.Warn(fmt.Sprintf("value is not addressible. TIP: consider to return a pointer from %s", vertexVal.Value.Type()), zap.String("type", vertexVal.Value.Type().String()))
						c.logger.Warn("making a fresh pointer")
						nt := reflect.New(vertexVal.Value.Type())
						inInterface = append(inInterface, nt)
					}
				}

				err := c.callDependerFns(vertex, inInterface)
				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func (c *Cascade) traverseCallDependers(vertex *structures.Vertex) error {
	in := make([]reflect.Value, 0, 2)
	// add service itself
	in = append(in, reflect.ValueOf(vertex.Iface))

	for i := 0; i < len(vertex.Meta.DepsList); i++ {
		// get dependency id (vertex id)
		depID := vertex.Meta.DepsList[i].Name
		// find vertex which provides dependency
		providers := c.graph.FindProviders(depID)
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

	err := c.callDependerFns(vertex, in)
	if err != nil {
		return err
	}

	return nil
}

func (c *Cascade) callDependerFns(vertex *structures.Vertex, in []reflect.Value) error {
	//type implements Depender interface
	if reflect.TypeOf(vertex.Iface).Implements(reflect.TypeOf((*Depender)(nil)).Elem()) {
		// if type implements Depender() it should has FnsProviderToInvoke
		if vertex.Meta.DepsList != nil {
			for k := 0; k < len(vertex.Meta.FnsDependerToInvoke); k++ {
				m, ok := reflect.TypeOf(vertex.Iface).MethodByName(vertex.Meta.FnsDependerToInvoke[k])
				if !ok {
					c.logger.Error("type has missing method in FnsDependerToInvoke", zap.String("vertex id", vertex.ID), zap.String("method", vertex.Meta.FnsDependerToInvoke[k]))
					return errors.New("type has missing method in FnsDependerToInvoke")
				}

				ret := m.Func.Call(in)
				// handle error
				if len(ret) > 0 {
					// error is the last return parameter
					rErr := ret[len(ret)-1].Interface()
					if rErr != nil {
						if e, ok := rErr.(error); ok && e != nil {
							c.logger.Error("error calling Registers", zap.String("vertex id", vertex.ID), zap.Error(e))
							return e
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

func (c *Cascade) findInitParameters(vertex *structures.Vertex) ([]reflect.Value, error) {
	in := make([]reflect.Value, 0, 2)

	// add service itself
	in = append(in, reflect.ValueOf(vertex.Iface))

	// add dependencies
	if len(vertex.Meta.InitDepsList) > 0 {
		for i := 0; i < len(vertex.Meta.InitDepsList); i++ {
			depID := vertex.Meta.InitDepsList[i].Name
			v := c.graph.FindProviders(depID)
			var err error
			in, err = c.traverseProviders(vertex.Meta.InitDepsList[i], v[0], depID, vertex.ID, in)
			if err != nil {
				return nil, err
			}
		}
	}
	return in, nil
}

func (c *Cascade) traverseProviders(depsEntry structures.DepsEntry, depVertex *structures.Vertex, depID string, calleeID string, in []reflect.Value) ([]reflect.Value, error) {
	// we need to call all providers first
	err := c.traverseCallProvider(depVertex, []reflect.Value{reflect.ValueOf(depVertex.Iface)}, calleeID)
	if err != nil {
		return nil, err
	}

	// to index function name in defer
	for providerID, val := range depVertex.Provides {
		if providerID == depID {
			switch {
			case *val.IsReference == *depsEntry.IsReference:
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
					c.logger.Warn(fmt.Sprintf("value is not addressible. TIP: consider to return a pointer from %s", val.Value.Type()), zap.String("type", val.Value.Type().String()))
					c.logger.Warn("making a fresh pointer")
					nt := reflect.New(val.Value.Type())
					in = append(in, nt)
				}
			}
		}
	}

	return in, nil
}

type TmpStr struct {
	N string
}

func (t TmpStr) Name() string {
	return t.N
}

func (c *Cascade) traverseCallProvider(v *structures.Vertex, in []reflect.Value, callerID string) error {
	// to index function name in defer
	i := 0
	defer func() {
		if r := recover(); r != nil {
			c.logger.Error("panic during the function call", zap.String("function name", v.Meta.FnsProviderToInvoke[i]), zap.String("error", fmt.Sprint(r)))
		}
	}()
	// type implements Provider interface
	if reflect.TypeOf(v.Iface).Implements(reflect.TypeOf((*Provider)(nil)).Elem()) {
		// if type implements Provider() it should has FnsProviderToInvoke
		if v.Meta.FnsProviderToInvoke != nil {
			// go over all function to invoke
			// invoke it
			// and save its return values
			for i = 0; i < len(v.Meta.FnsProviderToInvoke); i++ {
				m, ok := reflect.TypeOf(v.Iface).MethodByName(v.Meta.FnsProviderToInvoke[i])
				if !ok {
					c.logger.Panic("should implement the Provider interface", zap.String("function name", v.Meta.FnsProviderToInvoke[i]))
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

					callerV := c.graph.GetVertex(callerID)
					if callerV == nil {
						return errors.New("caller vertex is nil")
					}

					// skip function receiver
					for j := 1; j < m.Func.Type().NumIn(); j++ {
						// current function IN type (interface)
						t := m.Func.Type().In(j)
						if t.Kind() != reflect.Interface {
							c.logger.Panic("Provider accepts only interfaces", zap.String("function name", v.Meta.FnsProviderToInvoke[i]))
						}

						// if Caller struct implements interface -- ok, add it to the inCopy list
						// else panic
						if reflect.TypeOf(callerV.Iface).Implements(t) == false {
							c.logger.Panic("Caller should implement callee interface", zap.String("function name", v.Meta.FnsProviderToInvoke[i]))
						}

						inCopy = append(inCopy, reflect.ValueOf(callerV.Iface))
					}
				}

				ret := m.Func.Call(inCopy)
				// handle error
				if len(ret) > 1 {
					rErr := ret[1].Interface()
					if rErr != nil {
						if e, ok := rErr.(error); ok && e != nil {
							c.logger.Error("error occurred in the traverseCallProvider", zap.String("vertex id", v.ID))
							return e
						}
						return errUnknownErrorOccurred
					}

					// add the value to the Providers
					err := v.AddProvider(removePointerAsterisk(ret[0].Type().String()), ret[0], isReference(ret[0].Type()), in[0].Kind())
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

func (c *Cascade) callServeFn(vertex *structures.Vertex, in []reflect.Value) *result {
	m, _ := reflect.TypeOf(vertex.Iface).MethodByName(ServeMethodName)
	ret := m.Func.Call(in)
	res := ret[0].Interface()
	if res != nil {
		if e, ok := res.(chan error); ok && e != nil {
			return &result{
				errCh:    e,
				exit:     make(chan struct{}, 2),
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
func (c *Cascade) callConfigureFn(vertex *structures.Vertex, in []reflect.Value) error {
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

func (c *Cascade) callStopFn(vertex *structures.Vertex, in []reflect.Value) error {
	// Call Stop() method, which returns only error (or nil)
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

func (c *Cascade) stop(vID string) error {
	vertex := c.graph.GetVertex(vID)

	in := make([]reflect.Value, 0, 1)
	// add service itself
	in = append(in, reflect.ValueOf(vertex.Iface))

	err := c.callStopFn(vertex, in)
	if err != nil {
		c.logger.Error("error occurred during the stop", zap.String("vertex id", vertex.ID))
		return err
	}

	if reflect.TypeOf(vertex.Iface).Implements(reflect.TypeOf((*Graceful)(nil)).Elem()) {
		err = c.close(vertex.ID, in)
		if err != nil {
			c.logger.Error("error occurred during the close", zap.String("vertex id", vertex.ID))
			return err
		}
	}

	return nil
}

// TODO add stack to the all of the log events
func (c *Cascade) close(vID string, in []reflect.Value) error {
	v := c.graph.GetVertex(vID)
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

func (c *Cascade) sendExitSignal(sorted []*structures.Vertex) {
	for _, v := range sorted {
		// get result by vertex ID
		tmp := c.results[v.ID]
		if tmp == nil {
			continue
		}
		// send exit signal to the goroutine in sorted order
		c.logger.Debug("sending exit signal to the vertex from the main thread", zap.String("vertex id", tmp.vertexID))
		tmp.exit <- struct{}{}

		c.results[v.ID] = nil
	}
}

func (c *Cascade) sendResultToUser(res *result) {
	c.userResultsCh <- &Result{
		Error: Error{
			Err:   res.err,
			Code:  0,
			Stack: nil,
		},
		VertexID: res.vertexID,
	}
}

func (c *Cascade) shutdown(n *structures.DllNode) {
	nCopy := n
	for nCopy != nil {
		err := c.stop(nCopy.Vertex.ID)
		if err != nil {
			// TODO do not return until finished
			// just log the errors
			// stack it in slice and if slice is not empty, print it ??
			c.logger.Error("error occurred during the services stopping", zap.String("vertex id", nCopy.Vertex.ID), zap.Error(err))
		}
		if channel, ok := c.results[nCopy.Vertex.ID]; ok && channel != nil {
			channel.exit <- struct{}{}
		}

		// next DLL node
		nCopy = nCopy.Next
	}
}

// serve run configure (if exist) and callServeFn for each node and put the results in the map
func (c *Cascade) serve(n *structures.DllNode) error {
	// handle all configure
	in := make([]reflect.Value, 0, 1)
	// add service itself
	in = append(in, reflect.ValueOf(n.Vertex.Iface))

	res := c.callServeFn(n.Vertex, in)
	if res != nil {
		c.results[res.vertexID] = res
	} else {
		c.logger.Error("nil result returned from the vertex", zap.String("vertex id", n.Vertex.ID))
		return fmt.Errorf("nil result returned from the vertex, vertex id: %s", n.Vertex.ID)
	}

	c.poll(res)
	if c.restartedTime[n.Vertex.ID] != nil {
		*c.restartedTime[n.Vertex.ID] = time.Now()
	} else {
		tmp := time.Now()
		c.restartedTime[n.Vertex.ID] = &tmp
	}

	return nil
}

func (c *Cascade) checkLeafErrorTime(res *result) bool {
	return c.restartedTime[res.vertexID] != nil && c.restartedTime[res.vertexID].After(*c.errorTime[res.vertexID])
}

// poll is used to poll the errors from the vertex
// and exit from it
func (c *Cascade) poll(r *result) {
	rr := r
	go func(res *result) {
		for {
			select {
			// error
			case e := <-res.errCh:
				if e != nil {
					// set error time
					c.rwMutex.Lock()
					if c.errorTime[res.vertexID] != nil {
						*c.errorTime[res.vertexID] = time.Now()
					} else {
						tmp := time.Now()
						c.errorTime[res.vertexID] = &tmp
					}
					c.rwMutex.Unlock()

					c.logger.Error("error processed in poll", zap.String("vertex id", res.vertexID), zap.Error(e))

					// set the error
					res.err = e

					// send handleErrorCh signal
					c.handleErrorCh <- res
				}
			// exit from the goroutine
			case <-res.exit:
				c.rwMutex.Lock()
				c.logger.Info("got exit signal", zap.String("vertex id", res.vertexID))
				err := c.stop(res.vertexID)
				if err != nil {
					c.logger.Error("error during exit signal", zap.String("error while stopping the vertex:", res.vertexID), zap.Error(err))
					c.rwMutex.Unlock()
				}
				c.rwMutex.Unlock()
				return
			}
		}
	}(rr)
}

func (c *Cascade) register(name string, vertex interface{}, order int) error {
	// check the vertex
	if c.graph.HasVertex(name) {
		return errVertexAlreadyExists(name)
	}

	meta := structures.Meta{
		Order: order,
	}

	// just push the vertex
	// here we can append in future some meta information
	c.graph.AddVertex(name, vertex, meta)
	return nil
}

func (c *Cascade) backoffInit(v *structures.Vertex) func() error {
	return func() error {
		// we already checked the Interface satisfaction
		// at this step absence of Init() is impossible
		init, _ := reflect.TypeOf(v.Iface).MethodByName(InitMethodName)

		err := c.callInitFn(init, v)
		if err != nil {
			c.logger.Error("error occurred during the call INIT function", zap.String("vertex id", v.ID), zap.Error(err))
			return err
		}

		return nil
	}
}

func (c *Cascade) configure(n *structures.DllNode) error {
	// handle all configure
	in := make([]reflect.Value, 0, 1)
	// add service itself
	in = append(in, reflect.ValueOf(n.Vertex.Iface))

	//var res Result
	if reflect.TypeOf(n.Vertex.Iface).Implements(reflect.TypeOf((*Graceful)(nil)).Elem()) {
		err := c.callConfigureFn(n.Vertex, in)
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *Cascade) backoffConfigure(n *structures.DllNode) func() error {
	return func() error {
		// handle all configure
		in := make([]reflect.Value, 0, 1)
		// add service itself
		in = append(in, reflect.ValueOf(n.Vertex.Iface))

		//var res Result
		if reflect.TypeOf(n.Vertex.Iface).Implements(reflect.TypeOf((*Graceful)(nil)).Elem()) {
			err := c.callConfigureFn(n.Vertex, in)
			if err != nil {
				c.logger.Error("error configuring the vertex", zap.String("vertex id", n.Vertex.ID), zap.Error(err))
				return err
			}
		}

		return nil
	}
}

// TODO move to the interface?
func (c *Cascade) restart() error {
	c.handleErrorCh <- &result{
		internalExit: true,
	}

	c.rwMutex.Lock()
	defer c.rwMutex.Unlock()

	c.logger.Info("restarting the Cascade")
	n := c.runList.Head

	// shutdown, send exit signals to every user Serve() goroutine
	c.shutdown(n)

	// reset the run list to initial state
	c.runList.Reset()
	// reset all results
	c.results = make(map[string]*result)
	// reset error timings
	c.errorTime = make(map[string]*time.Time)
	// reset restarted timings
	c.restartedTime = make(map[string]*time.Time)

	// re-start main thread
	c.startMainThread()

	// call configure
	nCopy := c.runList.Head
	for nCopy != nil {
		err := c.configure(nCopy)
		if err != nil {
			c.logger.Error("backoff failed", zap.String("vertex id", nCopy.Vertex.ID), zap.Error(err))
			return err
		}

		nCopy = nCopy.Next
	}

	nCopy = c.runList.Head
	for nCopy != nil {
		err := c.serve(n)
		if err != nil {
			return err
		}
		nCopy = nCopy.Next
	}

	return nil
}
