package endure

import (
	"context"
	"reflect"
	"time"

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

func (e *Endure) shutdown(n *DllNode) {
	numOfVertices := len(e.graph.Vertices)
	if numOfVertices == 0 {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Hour)
	defer cancel()
	c := make(chan string)

	// used to properly exit
	// if the total number of vertices equal to the stopped, it means, that we stopped all
	stopped := 0

	go func() {
		// process all nodes one by one
		nCopy := n
		for nCopy != nil {
			// if vertex is disabled, just skip it, but send to the channel ID
			if nCopy.Vertex.IsDisabled == true {
				c <- nCopy.Vertex.ID
				nCopy = nCopy.Next
				continue
			}

			// if vertex is Uninitialized or already stopped
			if nCopy.Vertex.GetState() == Uninitialized || nCopy.Vertex.GetState() == Stopped {
				c <- nCopy.Vertex.ID
				nCopy = nCopy.Next
				continue
			}

			// if we have a running poller, exit from it
			tmp, ok := e.results.Load(nCopy.Vertex.ID)
			if ok {
				channel := tmp.(*result)

				// exit from vertex poller
				channel.signal <- notify{}
				e.results.Delete(nCopy.Vertex.ID)
			}

			// call Stop on the Vertex
			err := e.internalStop(nCopy.Vertex.ID)
			if err != nil {
				c <- nCopy.Vertex.ID
				e.logger.Error("error stopping vertex", zap.String("vertex id", nCopy.Vertex.ID), zap.Error(err))
				nCopy = nCopy.Next
				continue
			}

			c <- nCopy.Vertex.ID
			nCopy.Vertex.SetState(Stopped)
			nCopy = nCopy.Next
		}
	}()

	for {
		select {
		// get notification about stopped vertex
		case vid := <-c:
			e.logger.Info("vertex stopped", zap.String("vertex id", vid))
			stopped += 1
			if stopped == numOfVertices {
				return
			}
		case <-ctx.Done():
			e.logger.Info("timeout exceed, some vertices are not stopped", zap.Error(ctx.Err()))
			return
		}
	}
}
