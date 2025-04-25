package endure

import (
	"reflect"
	"sort"

	"github.com/roadrunner-server/endure/v2/graph"
	"github.com/roadrunner-server/errors"
	"go.uber.org/zap"
)

func (e *Endure) serve() error {
	/*
		topological order
	*/
	vertices := e.graph.TopologicalOrder()
	serveVertices := make([]*graph.Vertex, len(vertices))
	copy(serveVertices, vertices)

	sort.Slice(serveVertices, func(i, j int) bool {
		return serveVertices[i].Weight() > serveVertices[j].Weight()
	})

	if len(serveVertices) == 0 {
		return errors.E(errors.Str("error occurred, nothing to run"))
	}

	for i := range serveVertices {
		if !serveVertices[i].IsActive() {
			continue
		}

		if !reflect.TypeOf(serveVertices[i].Plugin()).Implements(reflect.TypeOf((*Service)(nil)).Elem()) {
			continue
		}

		serveMethod, _ := reflect.TypeOf(serveVertices[i].Plugin()).MethodByName(ServeMethodName)

		var inVals []reflect.Value
		inVals = append(inVals, reflect.ValueOf(serveVertices[i].Plugin()))

		e.log.Debug("calling serve method", zap.String("plugin", serveVertices[i].ID().String()))

		ret := serveMethod.Func.Call(inVals)[0].Interface()
		if ret != nil {
			if errCh, ok := ret.(chan error); ok && errCh != nil {
				// check if we have an error in the user's channel
				select {
				case er := <-errCh:
					return errors.E(
						errors.FunctionCall,
						errors.Errorf(
							"serve error from the plugin %s stopping execution, error: %v",
							serveVertices[i].ID().String(), er),
					)
				default:
					// if we don't have an error in the user's channel, activate poller
					e.poll(&result{
						// listen for the user's error channel
						errCh:    errCh,
						vertexID: serveVertices[i].ID().String(),
					})
				}
			}
		}
	}

	return nil
}
