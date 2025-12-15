package endure

import (
	"context"
	stderr "errors"
	"reflect"
	"sync"

	"github.com/roadrunner-server/errors"
	"go.uber.org/zap"
)

func (e *Endure) stop() error {
	/*
		topological order
	*/
	vertices := e.graph.TopologicalOrder()

	if len(vertices) == 0 {
		return errors.E(errors.Str("error occurred, nothing to run"))
	}

	mu := new(sync.Mutex)
	errs := make([]error, 0, 2)
	wg := &sync.WaitGroup{}
	wg.Add(len(vertices))

	// reverse order
	for i := len(vertices) - 1; i >= 0; i-- {
		if !vertices[i].IsActive() {
			wg.Done()
			continue
		}

		if !reflect.TypeOf(vertices[i].Plugin()).Implements(reflect.TypeFor[Service]()) {
			wg.Done()
			continue
		}

		go func(i int) {
			defer wg.Done()
			stopMethod, _ := reflect.TypeOf(vertices[i].Plugin()).MethodByName(StopMethodName)

			var inVals []reflect.Value
			inVals = append(inVals, reflect.ValueOf(vertices[i].Plugin()))

			e.log.Debug(
				"calling stop function",
				zap.String("plugin", vertices[i].ID().String()),
			)

			ctx, cancel := context.WithTimeout(context.Background(), e.stopTimeout)
			inVals = append(inVals, reflect.ValueOf(ctx))

			ret := stopMethod.Func.Call(inVals)[0].Interface()
			if ret != nil {
				e.log.Error("failed to stop the plugin", zap.String("name", vertices[i].ID().String()), zap.Error(ret.(error)))
				mu.Lock()
				errs = append(errs, ret.(error))
				mu.Unlock()
			}

			cancel()
		}(i)
	}

	wg.Wait()

	if len(errs) > 0 {
		return stderr.Join(errs...)
	}

	return nil
}
