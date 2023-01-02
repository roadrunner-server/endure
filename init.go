package endure

import (
	"reflect"

	"github.com/roadrunner-server/errors"
	"golang.org/x/exp/slog"
)

func (e *Endure) init() error {
	/*
		topological order
	*/
	vertices := e.graph.TopologicalOrder()

	for i := 0; i < len(vertices); i++ {
		if !vertices[i].IsActive() {
			continue
		}

		initMethod, _ := reflect.TypeOf(vertices[i].Plugin()).MethodByName(InitMethodName)

		args := make([]reflect.Type, initMethod.Type.NumIn())
		for j := 0; j < initMethod.Type.NumIn(); j++ {
			args[j] = initMethod.Type.In(j)
		}

		var inVals []reflect.Value
		inVals = append(inVals, reflect.ValueOf(vertices[i].Plugin()))
		// has deps if > 1
		if len(args) > 1 {
			for j := 0; j < len(args[1:]); j++ {
				arg := args[1:][j]

				plugins := e.registar.Implements(arg)
				inVals = append(inVals, reflect.ValueOf(plugins[0].Plugin()))
			}
		}

		ret := initMethod.Func.Call(inVals)
		if len(ret) > 1 {
			return errors.E("Init function should return only error, `Init(args) error {}`")
		}

		if ret[0].Type() != reflect.TypeOf((*error)(nil)).Elem() {
			return errors.E("Init function return type should be the error")
		}

		if ret[0].Interface() != nil {
			// may panic here?
			if _, ok := ret[0].Interface().(error); !ok {
				return errors.E("Init function should return only error, `Init(args) error {}`")
			}

			if errors.Is(errors.Disabled, ret[0].Interface().(error)) {
				e.log.Debug(
					"plugin disabled",
					slog.String("name", vertices[i].ID().String()),
				)
				// delete vertex and continue
				plugins := e.graph.Remove(vertices[i].Plugin())

				for j := 0; j < len(plugins); j++ {
					e.log.Debug(
						"destination plugin disabled because root was disabled",
						slog.String("name", plugins[j].ID().String()),
					)
					e.registar.Remove(plugins[j].Plugin())
				}
				continue
			}

			return ret[0].Interface().(error)
		}

		if provider, ok := vertices[i].Plugin().(Provider); ok {
			out := provider.Provides()
			for j := 0; j < len(out); j++ {
				providesMethod, okk := reflect.TypeOf(vertices[i].Plugin()).MethodByName(out[j].Method)
				if !okk {
					e.log.Warn("registered method doesn't exists ??")
					continue
				}

				e.registar.Update(vertices[i].Plugin(), out[j].Type, func() reflect.Value {
					vals := providesMethod.Func.Call([]reflect.Value{inVals[0]})
					if len(vals) > 1 {
						panic(">1")
					}

					return vals[0]
				})
			}

			continue
		}
	}

	return nil
}
