package endure

import (
	"reflect"
)

func (e *Endure) stop() error {
	/*
		topological order
	*/
	vertices := e.graph.TopologicalOrder()

	// reverse order
	for i := len(vertices) - 1; i >= 0; i-- {
		if !vertices[i].IsActive() {
			continue
		}

		if !reflect.TypeOf(vertices[i].Plugin()).Implements(reflect.TypeOf((*Service)(nil)).Elem()) {
			continue
		}

		stopMethod, _ := reflect.TypeOf(vertices[i].Plugin()).MethodByName(StopMethodName)

		var inVals []reflect.Value
		inVals = append(inVals, reflect.ValueOf(vertices[i].Plugin()))

		ret := stopMethod.Func.Call(inVals)[0].Interface()
		if ret != nil {
			return ret.(error)
		}
	}

	return nil
}
