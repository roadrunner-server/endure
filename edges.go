package endure

import (
	"reflect"

	"github.com/roadrunner-server/endure/v2/graph"
	"github.com/roadrunner-server/errors"
	"go.uber.org/zap"
)

func (e *Endure) resolveCollectorEdges(plugin any) error {
	// vertexID string, vertex any same vertex
	collector := plugin.(Collector)

	// retrieve the needed dependencies via Collects
	inEntries := collector.Collects()

	for i := range inEntries {
		res := e.registar.ImplementsExcept(inEntries[i].Type, plugin)
		if len(res) > 0 {
			for j := range res {
				e.graph.AddEdge(graph.CollectsConnection, res[j].Plugin(), plugin)
				e.log.Debug("collects edge found",
					zap.String("method", res[j].Method()),
					zap.String("src", e.graph.VertexById(res[j].Plugin()).ID().String()),
					zap.String("dest", e.graph.VertexById(plugin).ID().String()))
			}
		}
	}

	return nil
}

// resolveEdges adds edges between the vertices
// At this point, we know all plugins and all 'provides' values
func (e *Endure) resolveEdges() error {
	vertices := e.graph.Vertices()

	for i := range vertices {
		vertex := e.graph.VertexById(vertices[i].Plugin())
		initMethod, ok := vertex.ID().MethodByName(InitMethodName)
		if !ok {
			return errors.E("plugin should have the `Init(...) error` method")
		}

		args := make([]reflect.Type, initMethod.Type.NumIn())
		for j := range initMethod.Type.NumIn() {
			if isPrimitive(initMethod.Type.In(j).String()) {
				e.log.Error(
					"primitive type in the function parameters",
					zap.String("plugin", vertices[i].ID().String()),
					zap.String("type", initMethod.Type.In(j).String()),
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

		// we need to have the same number of plugins which implements the needed dep
		count := 0
		if len(args) > 1 {
			for j := 1; j < len(args); j++ {
				res := e.registar.ImplementsExcept(args[j], vertices[i].Plugin())
				if len(res) > 0 {
					count += 1
					for k := range res {
						// add graph edge
						e.graph.AddEdge(graph.InitConnection, res[k].Plugin(), vertex.Plugin())
						// log
						e.log.Debug(
							"init edge found",
							zap.Any("src", e.graph.VertexById(res[k].Plugin()).ID().String()),
							zap.Any("dest", e.graph.VertexById(vertex.Plugin()).ID().String()),
						)
					}
				}
			}

			// we should have here exactly the same number of the deps implementing every particular arg
			if count != len(args[1:]) {
				// if there are no plugins that implement Init deps, remove this vertex from the tree
				del := e.graph.Remove(vertices[i].Plugin())
				for k := range del {
					e.registar.Remove(del[k].Plugin())
					e.log.Debug(
						"plugin disabled, not enough Init dependencies",
						zap.String("name", del[k].ID().String()),
					)
				}

				continue
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

	e.graph.TopologicalSort()

	// to notify user about the disabled plugins
	// after topological sorting, we remove all plugins with indegree > 0, because there are no edges to them
	if len(e.graph.TopologicalOrder()) != len(e.graph.Vertices()) {
		tpl := e.graph.TopologicalOrder()
		vrt := e.graph.Vertices()

		tmpM := make(map[string]struct{}, 2)
		for _, v := range tpl {
			tmpM[v.ID().String()] = struct{}{}
		}

		for _, v := range vrt {
			if _, ok := tmpM[v.ID().String()]; !ok {
				e.log.Warn("topological sort, plugin disabled", zap.String("plugin", v.ID().String()))
			}
		}
	}

	return nil
}
