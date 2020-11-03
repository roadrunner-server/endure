// +build !windows

package endure

import (
	"github.com/goccy/go-graphviz"
	"github.com/spiral/errors"
)

func (e *Endure) Visualize(vertices []*Vertex) error {
	const op = errors.Op("print_graph")
	gr := graphviz.New()
	graph, err := gr.Graph()
	if err != nil {
		return errors.E(op, err)
	}

	for i := 0; i < len(vertices); i++ {
		if len(vertices[i].Dependencies) > 0 {
			for j := 0; j < len(vertices[i].Dependencies); j++ {
				n, err := graph.CreateNode(vertices[i].ID)
				if err != nil {
					return errors.E(op, err)
				}

				m, err := graph.CreateNode(vertices[i].Dependencies[j].ID)
				if err != nil {
					return errors.E(op, err)
				}

				e, err := graph.CreateEdge("", n, m)
				if err != nil {
					return errors.E(op, err)
				}
				e.SetLabel("")
			}
		}
	}

	// write to file directly
	if err := gr.RenderFilename(graph, graphviz.PNG, "./graph.png"); err != nil {
		return errors.E(op, err)
	}
	return nil
}
