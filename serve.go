package endure

import (
	"reflect"
	"sort"

	"github.com/roadrunner-server/endure/v2/graph"
	"github.com/roadrunner-server/errors"
	"golang.org/x/exp/slog"
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

	for i := 0; i < len(serveVertices); i++ {
		if !serveVertices[i].IsActive() {
			continue
		}

		if !reflect.TypeOf(serveVertices[i].Plugin()).Implements(reflect.TypeOf((*Service)(nil)).Elem()) {
			continue
		}

		serveMethod, _ := reflect.TypeOf(serveVertices[i].Plugin()).MethodByName(ServeMethodName)

		var inVals []reflect.Value
		inVals = append(inVals, reflect.ValueOf(serveVertices[i].Plugin()))

		e.log.Debug("calling serve method", slog.String("plugin", serveVertices[i].ID().String()))

		ret := serveMethod.Func.Call(inVals)[0].Interface()
		if ret != nil {
			if err, ok := ret.(chan error); ok && err != nil {
				// error come right after we start serving the vertex
				if len(err) > 0 {
					// read the error
					err := <-err
					return errors.E(
						errors.FunctionCall,
						errors.Errorf(
							"got initial serve error from the Vertex %s, stopping execution, error: %v",
							serveVertices[i].ID().String(), err),
					)
				}
				e.poll(&result{
					errCh:    err,
					signal:   make(chan notify),
					vertexID: serveVertices[i].ID().String(),
				})
			}
		}
	}

	return nil
}
