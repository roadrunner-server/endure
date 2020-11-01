package endure

import (
	"reflect"

	"github.com/cenkalti/backoff/v4"
	"github.com/spiral/errors"
	"go.uber.org/zap"
)

func (e *Endure) sendResultToUser(res *result) {
	e.userResultsCh <- &Result{
		Error:    res.err,
		VertexID: res.vertexID,
	}
}

// traverseBackStop used to visit every Prev node and internalStop vertices
func (e *Endure) traverseBackStop(n *DllNode) {
	const op = errors.Op("traverse_back_stop")
	e.logger.Debug("stopping vertex in the first Serve call", zap.String("vertex id", n.Vertex.ID))
	nCopy := n
	for nCopy != nil {
		err := e.internalStop(nCopy.Vertex.ID)
		if err != nil {
			// ignore errors from internal_stop
			e.logger.Error("failed to traverse vertex back", zap.String("vertex id", nCopy.Vertex.ID), zap.Error(errors.E(op, err)))
		}
		nCopy = nCopy.Prev
	}
}

func (e *Endure) retryHandler(res *result) {
	const op = errors.Op("internal_retry_handler")
	// get vertex from the graph
	vertex := e.graph.GetVertex(res.vertexID)
	if vertex == nil {
		e.logger.Error("failed to get vertex from the graph, vertex is nil", zap.String("vertex id from the handleErrorCh channel", res.vertexID))
		e.userResultsCh <- &Result{
			Error:    errors.E(op, errors.Traverse, errors.Str("failed to get vertex from the graph, vertex is nil")),
			VertexID: res.vertexID,
		}
		return
	}

	// reset vertex and dependencies to the initial state
	// numOfDeps and visited/visiting
	vertices := e.graph.Reset(vertex)

	// Topologically sort the graph
	sorted, err := TopologicalSort(vertices)
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
	// internal_stop will be called inside poller
	e.sendStopSignal(sorted)

	// Init backoff
	b := backoff.NewExponentialBackOff()
	b.MaxElapsedTime = e.maxInterval
	b.InitialInterval = e.initialInterval

	affectedRunList := NewDoublyLinkedList()
	for i := 0; i <= len(sorted)-1; i++ {
		affectedRunList.Push(sorted[i])
	}

	// call internal_init
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

	// call serveInternal
	headCopy = affectedRunList.Head
	for headCopy != nil {
		err = e.serveInternal(headCopy)
		if err != nil {
			e.userResultsCh <- &Result{
				Error:    errors.E(op, errors.FunctionCall, errors.Errorf("error during the Serve function call")),
				VertexID: headCopy.Vertex.ID,
			}
			e.logger.Error("fatal error during the serveInternal in the main thread", zap.String("vertex id", headCopy.Vertex.ID), zap.Error(err))
			return
		}
		headCopy = headCopy.Next
	}

	e.sendResultToUser(res)
}

func (e *Endure) backoffInit(v *Vertex) func() error {
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
