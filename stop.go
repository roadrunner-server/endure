package endure

import (
	"context"
	"reflect"
	"time"

	"github.com/spiral/endure/structures"
	"github.com/spiral/errors"
	"go.uber.org/zap"
)

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
	e.logger.Debug("calling stop function on the vertex", zap.String("vertex id", vertex.ID))
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

func (e *Endure) shutdown(n *structures.DllNode) {
	// channel with nodes to stop
	sh := make(chan *structures.DllNode)
	// todo remove magic time const
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
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
				// stack it in slice and if slice is not empty, visualize it ??
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
