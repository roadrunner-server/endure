package endure

import (
	"context"
	stderr "errors"
	"log/slog"
	"reflect"
	"sync"

	"golang.org/x/sync/semaphore"

	"github.com/roadrunner-server/errors"
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
	sema := semaphore.NewWeighted(int64(len(vertices)))

	// reverse order
	for i := len(vertices) - 1; i >= 0; i-- {
		if !vertices[i].IsActive() {
			continue
		}

		if !reflect.TypeOf(vertices[i].Plugin()).Implements(reflect.TypeOf((*Service)(nil)).Elem()) {
			continue
		}

		_ = sema.Acquire(context.Background(), 1)
		go func(i int) {
			stopMethod, _ := reflect.TypeOf(vertices[i].Plugin()).MethodByName(StopMethodName)

			var inVals []reflect.Value
			inVals = append(inVals, reflect.ValueOf(vertices[i].Plugin()))

			e.log.Debug(
				"calling stop function",
				slog.String("plugin", vertices[i].ID().String()),
			)

			ctx, cancel := context.WithTimeout(context.Background(), e.stopTimeout)
			inVals = append(inVals, reflect.ValueOf(ctx))

			ret := stopMethod.Func.Call(inVals)[0].Interface()
			if ret != nil {
				e.log.Error("failed to stop the plugin", slog.Any("error", ret.(error)))
				mu.Lock()
				errs = append(errs, ret.(error))
				mu.Unlock()
			}

			sema.Release(1)
			cancel()
		}(i)
	}

	_ = sema.Acquire(context.Background(), int64(len(vertices)))

	if len(errs) > 0 {
		return stderr.Join(errs...)
	}

	return nil
}
