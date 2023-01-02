package endure

import (
	"reflect"

	"github.com/roadrunner-server/endure/v2/graph"
	"github.com/roadrunner-server/errors"
	"golang.org/x/exp/slog"
)

func (e *Endure) resolveCollectorEdges(plugin any) error {
	// vertexID string, vertex any same vertex
	collector := plugin.(Collector)

	// retrieve the needed dependencies via Collects
	inEntries := collector.Collects()

	for i := 0; i < len(inEntries); i++ {
		implList := e.registar.Implements(inEntries[i].Type)
		if len(implList) > 0 {
			for j := 0; j < len(implList); j++ {
				e.graph.AddEdge(graph.CollectsConnection, implList[j].Plugin(), plugin)
				/*
					Here we need to init the
				*/
				e.log.Debug("collects edge found",
					slog.Any("methods", implList[j].Method()),
					slog.Any("src", e.graph.VertexById(implList[j].Plugin()).ID().String()),
					slog.Any("dest", e.graph.VertexById(plugin).ID().String()))
			}
		}
	}

	return nil
}

// resolveEdges adds edges between the vertices
// At this point, we know all plugins and all provides values
func (e *Endure) resolveEdges() error {
	vertices := e.graph.Vertices()

	for i := 0; i < len(vertices); i++ {
		vertex := e.graph.VertexById(vertices[i].Plugin())
		initMethod, ok := vertex.ID().MethodByName(InitMethodName)
		if !ok {
			return errors.E("plugin should have the `Init(...) error` method")
		}

		args := make([]reflect.Type, initMethod.Type.NumIn())
		for j := 0; j < initMethod.Type.NumIn(); j++ {
			if isPrimitive(initMethod.Type.In(j).String()) {
				e.log.Error(
					"primitive type in the function parameters",
					nil,
					slog.String("plugin", vertices[i].ID().String()),
					slog.String("type", initMethod.Type.In(j).String()),
				)

				return errors.E("Init method should not receive primitive types (like string, int, etc). It should receive only interfaces")
			}

			// check kind only for the 1..n In types (0-th is always receiver)
			if j > 0 {
				if initMethod.Type.In(j).Kind() != reflect.Interface {
					return errors.E("argument passed to the Init should be of the Interface type: e.g: func(p *Plugin) Init(io.Writer), not func(p *Plugin) Init(SomeStructure)")
				}
			}

			args[j] = initMethod.Type.In(j)
		}

		if len(args) > 1 {
			res := e.registar.Implements(args[1:]...)
			if len(res) != 0 {
				for j := 0; j < len(res); j++ {
					// add graph edge
					e.graph.AddEdge(graph.InitConnection, res[j].Plugin(), vertex.Plugin())
					// log
					e.log.Debug(
						"init edge found",
						slog.Any("src", e.graph.VertexById(res[j].Plugin()).ID().String()),
						slog.Any("dest", e.graph.VertexById(vertex.Plugin()).ID().String()),
					)
				}
			}
		}

		// we don't have a collector() method
		if _, okc := vertices[i].Plugin().(Collector); !okc {
			continue
		}

		err := e.resolveCollectorEdges(vertices[i].Plugin())
		if err != nil {
			return err
		}
	}

	ok := e.graph.TopologicalSort()
	if !ok {
		return errors.E("cyclic dependencies found, see the DEBUG log")
	}

	return nil
}
