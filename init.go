package endure

import (
	"log/slog"
	"reflect"

	"github.com/roadrunner-server/errors"
)

func (e *Endure) init() error {
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

		initMethod, _ := reflect.TypeOf(vertices[i].Plugin()).MethodByName(InitMethodName)

		args := make([]reflect.Type, initMethod.Type.NumIn())
		for j := 0; j < initMethod.Type.NumIn(); j++ {
			args[j] = initMethod.Type.In(j)
		}

		var inVals []reflect.Value
		inVals = append(inVals, reflect.ValueOf(vertices[i].Plugin()))
		// has deps if > 1
		if len(args) > 1 {
			// exclude first arg (it's receiver)
			arg := args[1:]
			for j := 0; j < len(arg); j++ {
				plugin := e.registar.ImplementsExcept(arg[j], vertices[i].Plugin())
				if len(plugin) == 0 {
					del := e.graph.Remove(vertices[i].Plugin())
					for k := 0; k < len(del); k++ {
						e.registar.Remove(del[k].Plugin())
						e.log.Debug(
							"plugin disabled, not enough Init dependencies",
							slog.String("name", del[k].ID().String()),
						)
					}

					continue
				}

				// check if the provided plugin dep has a method
				// existence of the method indicates, that the dep provided by this plugin should be obtained via the method call
				switch plugin[0].Method() == "" {
				// we don't have a method, that means, plugin itself implements the dep
				case true:
					inVals = append(inVals, reflect.ValueOf(plugin[0].Plugin()))

					// we have a method, thus we need to get the value, because previous plugin have registered it's provided deps
				case false:
					value, ok := e.registar.TypeValue(plugin[0].Plugin(), arg[j])
					if !ok {
						return errors.E("this is likely a bug, nil value from the implements. Value should be initialized due to the topological order")
					}
					inVals = append(inVals, value)
				}
			}
		}

		ret := initMethod.Func.Call(inVals)
		if len(ret) > 1 {
			// fatal error, clean the graph
			e.graph.Clean()
			return errors.E("Init function should return only error, `Init(args) error {}`")
		}

		if ret[0].Type() != reflect.TypeOf((*error)(nil)).Elem() {
			// fatal error, clean the graph
			e.graph.Clean()
			return errors.E("Init function return type should be the error")
		}

		if ret[0].Interface() != nil {
			// may panic here?
			if _, ok := ret[0].Interface().(error); !ok {
				// fatal error, clean the graph
				e.graph.Clean()
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

			// fatal error, clean the graph
			e.graph.Clean()
			return ret[0].Interface().(error)
		}

		// add vertex itself
		vrtx := vertices[i].Plugin()
		e.registar.Update(vrtx, reflect.TypeOf(vrtx), func() reflect.Value {
			return reflect.ValueOf(vrtx)
		})

		if provider, ok := vertices[i].Plugin().(Provider); ok {
			out := provider.Provides()
			for j := 0; j < len(out); j++ {
				providesMethod, okk := reflect.TypeOf(vertices[i].Plugin()).MethodByName(out[j].Method)
				if !okk {
					e.log.Warn("registered method doesn't exists ??")
					continue
				}

				tp := out[j].Type
				pl := vertices[i].Plugin()
				in := []reflect.Value{inVals[0]}
				e.registar.Update(pl, tp, func() reflect.Value {
					vals := providesMethod.Func.Call(in)
					if len(vals) != 1 {
						panic("provides method should provide only 1 arg - structure")
					}

					return vals[0]
				})
			}
		}
	}

	inactive := 0
	for i := 0; i < len(vertices); i++ {
		if !vertices[i].IsActive() {
			inactive++
		}
	}

	if inactive == len(vertices) {
		return errors.E(errors.Str("All plugins are disabled, nothing to serve"))
	}

	return nil
}
