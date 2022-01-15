package endure

import (
	"reflect"

	"github.com/cenkalti/backoff/v4"
	"github.com/roadrunner-server/endure/pkg/fsm"
	"github.com/roadrunner-server/endure/pkg/graph"
	ll "github.com/roadrunner-server/endure/pkg/linked_list"
	"github.com/roadrunner-server/endure/pkg/vertex"
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
func (e *Endure) traverseBackStop(n *ll.DllNode) {
	const op = errors.Op("endure_traverse_back_stop")
	e.logger.Debug("stopping vertex in the first Serve call", zap.String("id", n.Vertex.ID))
	nCopy := n
	err := e.shutdown(nCopy, false)
	if err != nil {
		nCopy.Vertex.SetState(fsm.Error)
		// ignore errors from internal_stop
		e.logger.Error("failed to traverse vertex back", zap.String("id", nCopy.Vertex.ID), zap.Error(errors.E(op, err)))
	}
}

func (e *Endure) retryHandler(res *result) {
	const op = errors.Op("endure_retry_handler")
	// get vertex from the graph
	vrtx := e.graph.GetVertex(res.vertexID)
	if vrtx == nil {
		e.logger.Error("failed to get vertex from the graph, vertex is nil", zap.String("id from the handleErrorCh channel", res.vertexID))
		e.userResultsCh <- &Result{
			Error:    errors.E(op, errors.Traverse, errors.Str("failed to get vertex from the graph, vertex is nil")),
			VertexID: res.vertexID,
		}
		return
	}

	// stop without setting Stopped state to the Endure
	n := e.runList.Head
	err := e.shutdown(n, true)
	if err != nil {
		e.logger.Error("error happened during shutdown", zap.Error(err))
	}

	// reset vertex and dependencies to the initial state
	// numOfDeps and visited/visiting
	vertices := e.graph.Reset(vrtx)

	// Topologically sort the graph
	sorted, err := graph.TopologicalSort(vertices)
	if err != nil {
		e.logger.Error("error sorting the graph", zap.Error(err))
		return
	}
	if sorted == nil {
		e.logger.Error("sorted list should not be nil", zap.String("id from the handleErrorCh channel", res.vertexID))
		e.userResultsCh <- &Result{
			Error:    errors.E(op, errors.Traverse, errors.Str("failed to topologically sort the graph")),
			VertexID: res.vertexID,
		}
		return
	}

	// Init backoff
	b := backoff.NewExponentialBackOff()
	b.MaxElapsedTime = e.maxInterval
	b.InitialInterval = e.initialInterval

	affectedRunList := ll.NewDoublyLinkedList()
	for i := 0; i <= len(sorted)-1; i++ {
		affectedRunList.Push(sorted[i])
	}

	// call internal_init
	headCopy := affectedRunList.Head
	for headCopy != nil {
		berr := backoff.Retry(e.backoffInit(headCopy.Vertex), b)
		if berr != nil {
			e.logger.Error("backoff failed", zap.String("id", headCopy.Vertex.ID), zap.Error(berr))
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
		err := e.serveInternal(headCopy)
		if err != nil {
			e.userResultsCh <- &Result{
				Error:    errors.E(op, errors.FunctionCall, errors.Errorf("error during the Serve function call")),
				VertexID: headCopy.Vertex.ID,
			}
			e.logger.Error("fatal error during the serveInternal in the main thread", zap.String("id", headCopy.Vertex.ID), zap.Error(err))
			return
		}
		headCopy = headCopy.Next
	}

	e.sendResultToUser(res)
}

func (e *Endure) backoffInit(v *vertex.Vertex) func() error {
	return func() error {
		const op = errors.Op("endure_backoff_init")
		// we already checked the Interface satisfaction
		// at this step absence of Init() is impossible
		init, _ := reflect.TypeOf(v.Iface).MethodByName(InitMethodName)
		v.SetState(fsm.Initializing)
		err := e.callInitFn(init, v)
		if err != nil {
			v.SetState(fsm.Error)
			e.logger.Error("error occurred during the call INIT function", zap.String("id", v.ID), zap.Error(err))
			return errors.E(op, errors.FunctionCall, err)
		}

		v.SetState(fsm.Initialized)
		return nil
	}
}
