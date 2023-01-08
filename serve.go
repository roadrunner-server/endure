package endure

import (
	"reflect"

	"github.com/roadrunner-server/errors"
)

func (e *Endure) serve() error {
	/*
		topological order
	*/
	vertices := e.graph.TopologicalOrder()

	if len(vertices) == 0 {
		return errors.E(errors.Str("error occurred, nothing to run"))
	}

	for i := 0; i < len(vertices); i++ {
		if !vertices[i].IsActive() {
			continue
		}

		if !reflect.TypeOf(vertices[i].Plugin()).Implements(reflect.TypeOf((*Service)(nil)).Elem()) {
			continue
		}

		serveMethod, _ := reflect.TypeOf(vertices[i].Plugin()).MethodByName(ServeMethodName)

		var inVals []reflect.Value
		inVals = append(inVals, reflect.ValueOf(vertices[i].Plugin()))

		ret := serveMethod.Func.Call(inVals)[0].Interface()
		if ret != nil {
			if err, ok := ret.(chan error); ok && err != nil {
				// error come right after we start serving the vrtx
				if len(err) > 0 {
					// read the error
					err := <-err
					return errors.E(
						errors.FunctionCall,
						errors.Errorf("got initial serve error from the Vertex %s, stopping execution, error: %v",
							vertices[i].ID().String(), err),
					)
				}
				e.poll(&result{
					errCh:    err,
					signal:   make(chan notify),
					vertexID: vertices[i].ID().String(),
				})
			}
		}
	}

	return nil
}
