package endure

import (
	"context"
	"reflect"

	"github.com/roadrunner-server/errors"
	"golang.org/x/exp/slog"
)

func (e *Endure) stop() error {
	/*
		topological order
	*/
	vertices := e.graph.TopologicalOrder()

	if len(vertices) == 0 {
		return errors.E(errors.Str("error occurred, nothing to run"))
	}

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

		e.log.Debug(
			"calling stop function",
			slog.String("plugin", vertices[i].ID().String()),
		)

		ctx, cancel := context.WithTimeout(context.Background(), e.stopTimeout)
		inVals = append(inVals, reflect.ValueOf(ctx))

		go func() {

		}()
		ret := stopMethod.Func.Call(inVals)[0].Interface()
		if ret != nil {
			cancel()
			return ret.(error)
		}

		cancel()
	}

	return nil
}
