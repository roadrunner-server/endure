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
func (c *Cascade) init(v *structures.Vertex) error {
	// we already checked the Interface satisfaction
	// at this step absence of Init() is impoosssibruuu
	init, _ := reflect.TypeOf(v.Iface).MethodByName(InitMethodName)

	err := c.callInitFn(init, v)
	if err != nil {
		c.logger.Error("error occurred during the call INIT function", zap.String("vertex id", v.Id), zap.Error(err))
		return err
	}

	return nil
}

func (c *Cascade) callInitFn(init reflect.Method, vertex *structures.Vertex) error {
	in := c.findInitParameters(vertex)

	// Iterate over dependencies
	// And search in Vertices for the provided types
	ret := init.Func.Call(in)
	rErr := ret[0].Interface()
	if rErr != nil {
		if e, ok := rErr.(error); ok && e != nil {
			c.logger.Error("error calling init", zap.String("vertex id", vertex.Id), zap.Error(e))
			return e
		} else {
			return unknownErrorOccurred
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
		err := vertex.AddProvider(removePointerAsterisk(in[0].Type().String()), in[0], isReference(in[0].Type()))
		if err != nil {
			return err
		}
		c.logger.Debug("value added successfully", zap.String("vertex id", vertex.Id), zap.String("parameter", in[0].Type().String()))

	} else {
		c.logger.Error("0 or less parameters for Init", zap.String("vertex id", vertex.Id))
		return errors.New("0 or less parameters for Init")
	}

	err := c.traverseCallProvider(vertex, []reflect.Value{reflect.ValueOf(vertex.Iface)})
	if err != nil {
		return err
	}

	err = c.traverseCallRegisters(vertex)
	if err != nil {
		return err
	}

	return nil
}

func (c *Cascade) traverseCallRegisters(vertex *structures.Vertex) error {
	in := make([]reflect.Value, 0, 1)

	// add service itself
	in = append(in, reflect.ValueOf(vertex.Iface))

	// add dependencies
	if len(vertex.Meta.DepsList) > 0 {
		for i := 0; i < len(vertex.Meta.DepsList); i++ {
			// get dependency id (vertex id)
			depId := vertex.Meta.DepsList[i].Name
			// find vertex which provides dependency
			v := c.graph.FindProvider(depId)

			in = c.traverseProviders(vertex.Meta.DepsList, v, depId, i, in)
		}
	}

	//type implements Depender interface
	if reflect.TypeOf(vertex.Iface).Implements(reflect.TypeOf((*Depender)(nil)).Elem()) {
		// if type implements Depender() it should has FnsProviderToInvoke
		if vertex.Meta.DepsList != nil {
			for i := 0; i < len(vertex.Meta.FnsDependerToInvoke); i++ {
				m, ok := reflect.TypeOf(vertex.Iface).MethodByName(vertex.Meta.FnsDependerToInvoke[i])
				if !ok {
					c.logger.Error("type has missing method in FnsDependerToInvoke", zap.String("vertex id", vertex.Id), zap.String("method", vertex.Meta.FnsDependerToInvoke[i]))
					return errors.New("type has missing method in FnsDependerToInvoke")
				}

				ret := m.Func.Call(in)
				// handle error
				if len(ret) > 0 {
					rErr := ret[0].Interface()
					if rErr != nil {
						if e, ok := rErr.(error); ok && e != nil {
							c.logger.Error("error calling Registers", zap.String("vertex id", vertex.Id), zap.Error(e))
							return e
						} else {
							return unknownErrorOccurred
						}
					}
				} else {
					return errors.New("depender should return Value and error types")
				}
			}
		}
	}
	return nil
}

func (c *Cascade) findInitParameters(vertex *structures.Vertex) []reflect.Value {
	in := make([]reflect.Value, 0, 2)

	// add service itself
	in = append(in, reflect.ValueOf(vertex.Iface))

	// add dependencies
	if len(vertex.Meta.InitDepsList) > 0 {
		for i := 0; i < len(vertex.Meta.InitDepsList); i++ {
			depId := vertex.Meta.InitDepsList[i].Name
			v := c.graph.FindProvider(depId)

			in = c.traverseProviders(vertex.Meta.InitDepsList, v, depId, i, in)
		}
	}
	return in
}

func (c *Cascade) traverseProviders(list []structures.DepsEntry, depVertex *structures.Vertex, depId string, i int, in []reflect.Value) []reflect.Value {
	for vertexId, val := range depVertex.Provides {
		if vertexId == depId {
			// value - reference and init dep also reference
			if *val.IsReference == *list[i].IsReference {
				in = append(in, *val.Value)
			} else if *val.IsReference {
				// same type, but difference in the refs
				// Init needs to be a value
				// But Vertex provided reference

				in = append(in, val.Value.Elem())
			} else if !*val.IsReference {
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
	return in
}

func (c *Cascade) traverseCallProvider(v *structures.Vertex, in []reflect.Value) error {
	// type implements Provider interface
	if reflect.TypeOf(v.Iface).Implements(reflect.TypeOf((*Provider)(nil)).Elem()) {
		// if type implements Provider() it should has FnsProviderToInvoke
		if v.Meta.FnsProviderToInvoke != nil {
			// go over all function to invoke
			// invoke it
			// and save its return values
			for i := 0; i < len(v.Meta.FnsProviderToInvoke); i++ {
				m, ok := reflect.TypeOf(v.Iface).MethodByName(v.Meta.FnsProviderToInvoke[i])
				if !ok {
					panic("method Provides should be")
				}

				ret := m.Func.Call(in)
				// handle error
				if len(ret) > 1 {
					rErr := ret[1].Interface()
					if rErr != nil {
						if e, ok := rErr.(error); ok && e != nil {
							c.logger.Error("error occurred in the traverseCallProvider", zap.String("vertex id", v.Id))
							return e
						} else {
							return unknownErrorOccurred
						}
					}

					// add the value to the Providers
					err := v.AddProvider(removePointerAsterisk(ret[0].Type().String()), ret[0], isReference(ret[0].Type()))
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
				vertexId: vertex.Id,
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
		return unknownErrorOccurred
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
		} else {
			return unknownErrorOccurred
		}
	}
	return nil

}

func (c *Cascade) stop(vId string) error {
	vertex := c.graph.GetVertex(vId)

	in := make([]reflect.Value, 0, 1)
	// add service itself
	in = append(in, reflect.ValueOf(vertex.Iface))

	err := c.callStopFn(vertex, in)
	if err != nil {
		c.logger.Error("error occurred during the stop", zap.String("vertex id", vertex.Id))
	}

	if reflect.TypeOf(vertex.Iface).Implements(reflect.TypeOf((*Graceful)(nil)).Elem()) {
		err = c.close(vertex.Id, in)
		if err != nil {
			c.logger.Error("error occurred during the close", zap.String("vertex id", vertex.Id))
		}
	}

	return nil
}

// TODO add stack to the all of the log events
func (c *Cascade) close(vId string, in []reflect.Value) error {
	v := c.graph.GetVertex(vId)
	// Call Close() method, which returns only error (or nil)
	m, _ := reflect.TypeOf(v.Iface).MethodByName(CloseMethodName)
	ret := m.Func.Call(in)
	rErr := ret[0].Interface()
	if rErr != nil {
		if e, ok := rErr.(error); ok && e != nil {
			return e
		} else {
			return unknownErrorOccurred
		}
	}
	return nil
}

func (c *Cascade) sendExitSignal(sorted []*structures.Vertex) {
	for _, v := range sorted {
		// get result by vertex ID
		tmp := c.results[v.Id]
		if tmp == nil {
			continue
		}
		// send exit signal to the goroutine in sorted order
		c.logger.Debug("sending exit signal to the vertex from the main thread", zap.String("vertex id", tmp.vertexId))
		tmp.exit <- struct{}{}

		c.results[v.Id] = nil
	}
}

func (c *Cascade) sendResultToUser(res *result) {
	c.userResultsCh <- &Result{
		Error: Error{
			Err:   res.err,
			Code:  0,
			Stack: nil,
		},
		VertexID: res.vertexId,
	}
}

func (c *Cascade) shutdown(n *structures.DllNode) {
	nCopy := n
	for nCopy != nil {
		err := c.stop(nCopy.Vertex.Id)
		if err != nil {
			// TODO do not return until finished
			// just log the errors
			// stack it in slice and if slice is not empty, print it ??
			c.logger.Error("error occurred during the services stopping", zap.String("vertex id", nCopy.Vertex.Id), zap.Error(err))
		}
		if channel, ok := c.results[nCopy.Vertex.Id]; ok && channel != nil {
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
		c.results[res.vertexId] = res
	} else {
		c.logger.Error("nil result returned from the vertex", zap.String("vertex id", n.Vertex.Id))
		return errors.New(fmt.Sprintf("nil result returned from the vertex, vertex id: %s", n.Vertex.Id))
	}

	c.poll(res)
	if c.restartedTime[n.Vertex.Id] != nil {
		*c.restartedTime[n.Vertex.Id] = time.Now()
	} else {
		tmp := time.Now()
		c.restartedTime[n.Vertex.Id] = &tmp
	}

	return nil
}

func (c *Cascade) checkLeafErrorTime(res *result) bool {
	return c.restartedTime[res.vertexId] != nil && (*c.restartedTime[res.vertexId]).After(*c.errorTime[res.vertexId])
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
					if c.errorTime[res.vertexId] != nil {
						*c.errorTime[res.vertexId] = time.Now()
					} else {
						tmp := time.Now()
						c.errorTime[res.vertexId] = &tmp
					}
					c.rwMutex.Unlock()

					c.logger.Error("error processed in poll", zap.String("vertex id", res.vertexId), zap.Error(e))

					// set the error
					res.err = e

					// send handleErrorCh signal
					c.handleErrorCh <- res
				}
			// exit from the goroutine
			case <-res.exit:
				c.rwMutex.Lock()
				c.logger.Info("got exit signal", zap.String("vertex id", res.vertexId))
				err := c.stop(res.vertexId)
				if err != nil {
					c.logger.Error("error during exit signal", zap.String("error while stopping the vertex:", res.vertexId), zap.Error(err))
					c.rwMutex.Unlock()
				}
				c.rwMutex.Unlock()
				return
			}
		}
	}(rr)
}

// TODO graph responsibility, not Cascade
func (c *Cascade) resetVertices(vertex *structures.Vertex) []*structures.Vertex {
	// restore number of dependencies for the root
	vertex.NumOfDeps = len(vertex.Dependencies)
	vertex.Visiting = false
	vertex.Visited = false
	vertices := make([]*structures.Vertex, 0, 5)
	vertices = append(vertices, vertex)

	tmp := make(map[string]*structures.Vertex)

	c.dfs(vertex.Dependencies, tmp)

	for _, v := range tmp {
		vertices = append(vertices, v)
	}
	return vertices
}

// TODO graph responsibility, not Cascade
func (c *Cascade) dfs(deps []*structures.Vertex, tmp map[string]*structures.Vertex) {
	for i := 0; i < len(deps); i++ {
		deps[i].Visited = false
		deps[i].Visiting = false
		deps[i].NumOfDeps = len(deps)
		tmp[deps[i].Id] = deps[i]
		c.dfs(deps[i].Dependencies, tmp)
	}
}

func (c *Cascade) register(name string, vertex interface{}) error {
	// check the vertex
	if c.graph.HasVertex(name) {
		return vertexAlreadyExists(name)
	}

	// just push the vertex
	// here we can append in future some meta information
	c.graph.AddVertex(name, vertex, structures.Meta{})
	return nil
}

func (c *Cascade) backoffInit(v *structures.Vertex) func() error {
	return func() error {
		// we already checked the Interface satisfaction
		// at this step absence of Init() is impossible
		init, _ := reflect.TypeOf(v.Iface).MethodByName(InitMethodName)

		err := c.callInitFn(init, v)
		if err != nil {
			c.logger.Error("error occurred during the call INIT function", zap.String("vertex id", v.Id), zap.Error(err))
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
				c.logger.Error("error configuring the vertex", zap.String("vertex id", n.Vertex.Id), zap.Error(err))
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
			c.logger.Error("backoff failed", zap.String("vertex id", nCopy.Vertex.Id), zap.Error(err))
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
