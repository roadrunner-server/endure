package endure

import (
	"github.com/roadrunner-server/endure/pkg/fsm"
	ll "github.com/roadrunner-server/endure/pkg/linked_list"
	"github.com/roadrunner-server/errors"
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
