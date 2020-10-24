// +build windows
package structures

import (
	"bytes"

	"github.com/goccy/go-graphviz"
)

func PrintGraph(vertices *Vertices) error {
	gr := graphviz.New()
	graph, err := gr.Graph()
	if err != nil {
		panic(err)
	}

	for i := 0; i < len(g.Vertices); i++ {
		if len(g.Vertices[i].Dependencies) > 0 {
			for j := 0; j < len(g.Vertices[i].Dependencies); j++ {
				n, err := graph.CreateNode(g.Vertices[i].ID)
				if err != nil {
					return err
				}

				m, err := graph.CreateNode(g.Vertices[i].Dependencies[j].ID)
				if err != nil {
					return err
				}

				e, err := graph.CreateEdge("", n, m)
				if err != nil {
					return err
				}
				e.SetLabel("")
			}
		}
	}

	var buf bytes.Buffer
	if err := gr.Render(graph, graphviz.PNG, &buf); err != nil {
		return err
	}

	// write to file directly
	if err := gr.RenderFilename(graph, graphviz.PNG, "./graph.png"); err != nil {
		return err
	}
	return nil
}
