package endure

import (
	"context"
	stderr "errors"
	"reflect"
	"slices"
	"sync"
	"sync/atomic"

	"github.com/roadrunner-server/endure/v2/graph"
	"github.com/roadrunner-server/errors"
	"go.uber.org/zap"
)

func (e *Endure) stop() error {
	vertices := e.graph.TopologicalOrder()
	if len(vertices) == 0 {
		return errors.E(errors.Str("error occurred, nothing to run"))
	}

	active := make([]*graph.Vertex, 0, len(vertices))
	activeSet := make(map[*graph.Vertex]struct{}, len(vertices))
	for _, v := range vertices {
		if !v.IsActive() {
			continue
		}
		active = append(active, v)
		activeSet[v] = struct{}{}
	}

	if len(active) == 0 {
		return nil
	}

	remaining := make(map[*graph.Vertex]*atomic.Int32, len(active))
	providers := make(map[*graph.Vertex][]*graph.Vertex, len(active))
	for _, u := range active {
		var deps int32
		for _, dest := range e.graph.ActiveDependents(u) {
			if _, ok := activeSet[dest]; !ok {
				continue
			}
			deps++
			providers[dest] = append(providers[dest], u)
		}
		cnt := &atomic.Int32{}
		cnt.Store(deps)
		remaining[u] = cnt
	}

	var (
		mu   sync.Mutex
		errs []error
		wg   sync.WaitGroup
	)
	wg.Add(len(active))

	serviceType := reflect.TypeFor[Service]()

	var stopVertex func(v *graph.Vertex)
	stopVertex = func(v *graph.Vertex) {
		defer wg.Done()

		if reflect.TypeOf(v.Plugin()).Implements(serviceType) {
			stopMethod, _ := reflect.TypeOf(v.Plugin()).MethodByName(StopMethodName)

			e.log.Debug(
				"calling stop function",
				zap.String("plugin", v.ID().String()),
			)

			ctx, cancel := context.WithTimeout(context.Background(), e.stopTimeout)
			ret := stopMethod.Func.Call([]reflect.Value{
				reflect.ValueOf(v.Plugin()),
				reflect.ValueOf(ctx),
			})[0].Interface()
			cancel()

			if ret != nil {
				err := ret.(error)
				e.log.Error("failed to stop the plugin", zap.String("name", v.ID().String()), zap.Error(err))
				mu.Lock()
				errs = append(errs, err)
				mu.Unlock()
			}
		}

		for _, provider := range providers[v] {
			if remaining[provider].Add(-1) == 0 {
				go stopVertex(provider)
			}
		}
	}

	for _, v := range slices.Backward(active) {
		if remaining[v].Load() == 0 {
			go stopVertex(v)
		}
	}

	wg.Wait()

	if len(errs) > 0 {
		return stderr.Join(errs...)
	}

	return nil
}
