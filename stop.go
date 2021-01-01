package endure

import (
	"context"
	"reflect"

	"github.com/spiral/errors"
	"go.uber.org/zap"
)

func (e *Endure) internalStop(vID string) error {
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

func (e *Endure) callStopFn(vertex *Vertex, in []reflect.Value) error {
	const op = errors.Op("internal_call_stop_function")
	// Call Stop() method, which returns only error (or nil)
	e.logger.Debug("calling internal_stop function on the vertex", zap.String("vertex id", vertex.ID))
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

// true -> next
// false -> prev
func (e *Endure) shutdown(n *DllNode, traverseNext bool) error {
	const op = errors.Op("shutdown")
	numOfVertices := calculateDepth(n, traverseNext)
	if numOfVertices == 0 {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), e.stopTimeout)
	defer cancel()
	c := make(chan string)

	// used to properly exit
	// if the total number of vertices equal to the stopped, it means, that we stopped all
	stopped := 0

	go func() {
		// process all nodes one by one
		nCopy := n
		for nCopy != nil {
			go func(v *Vertex) {
				// if vertex is disabled, just skip it, but send to the channel ID
				if v.IsDisabled == true {
					c <- v.ID
					return
				}

				// if vertex is Uninitialized or already stopped
				// Skip vertices which are not Started
				if v.GetState() != Started {
					c <- v.ID
					return
				}

				v.SetState(Stopping)

				// if we have a running poller, exit from it
				tmp, ok := e.results.Load(v.ID)
				if ok {
					channel := tmp.(*result)

					// exit from vertex poller
					channel.signal <- notify{}
					e.results.Delete(v.ID)
				}

				// call Stop on the Vertex
				err := e.internalStop(v.ID)
				if err != nil {
					v.SetState(Error)
					c <- v.ID
					e.logger.Error("error stopping vertex", zap.String("vertex id", v.ID), zap.Error(err))
					return
				}
				v.SetState(Stopped)
				c <- v.ID
			}(nCopy.Vertex)
			if traverseNext {
				nCopy = nCopy.Next
			} else {
				nCopy = nCopy.Prev
			}
		}
	}()

	for {
		select {
		// get notification about stopped vertex
		case vid := <-c:
			e.logger.Info("vertex stopped", zap.String("vertex id", vid))
			stopped++
			if stopped == numOfVertices {
				return nil
			}
		case <-ctx.Done():
			e.logger.Info("timeout exceed, some vertices are not stopped", zap.Error(ctx.Err()))
			// iterate to see vertices, which are not stopped
			VIDs := make([]string, 0, 1)
			for i := 0; i < len(e.graph.Vertices); i++ {
				state := e.graph.Vertices[i].GetState()
				if state == Started || state == Stopping {
					VIDs = append(VIDs, e.graph.Vertices[i].ID)
				}
			}
			if len(VIDs) > 0 {
				e.logger.Error("vertices which are not stopped", zap.Any("vertex id", VIDs))
			}

			return errors.E(op, errors.TimeOut, errors.Str("timeout exceed, some vertices may not be stopped and can cause memory leak"))
		}
	}
}

// Using to calculate number of Vertices in DLL
func calculateDepth(n *DllNode, traverse bool) int {
	num := 0
	if traverse {
		tmp := n
		for tmp != nil {
			num++
			tmp = tmp.Next
		}
		return num
	}
	tmp := n
	for tmp != nil {
		num++
		tmp = tmp.Prev
	}
	return num
}
