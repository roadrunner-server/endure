package graph

import (
	"fmt"
	"os"
	"reflect"
	"slices"
	"strings"
)

const (
	// InitMethodName is the function fn for the reflection
	InitMethodName = "Init"
)

// Graph manages the set of services and their edges
// type of the VerticesMap: directed
type Graph struct {
	// Map with vertices to have an easy access to it
	vertices map[reflect.Type]*Vertex
	// List of all Vertices
	topologicalOrder []*Vertex
}

// New initializes endure Graph
// According to the topological sorting, graph should be
// 1. DIRECTED
// 2. ACYCLIC
func New() *Graph {
	return &Graph{
		vertices:         make(map[reflect.Type]*Vertex),
		topologicalOrder: make([]*Vertex, 0),
	}
}

// HasVertex returns true or false if the vertex exists in the vertices map in the graph
func (g *Graph) HasVertex(plugin any) bool {
	tp := reflect.TypeOf(plugin)
	_, ok := g.vertices[tp]
	return ok
}

func (g *Graph) AddEdge(edgeType EdgeType, src, dest any) {
	e := &edge{
		src:            src,
		dest:           dest,
		connectionType: edgeType,
	}

	s := g.VertexById(e.src)
	d := g.VertexById(e.dest)

	s.edges = append(s.edges, e)
	d.indegree++
}

func (g *Graph) VertexById(plugin any) *Vertex {
	return g.vertices[reflect.TypeOf(plugin)]
}

func (g *Graph) Vertices() []*Vertex {
	v := make([]*Vertex, 0, len(g.vertices))

	for _, vrx := range g.vertices {
		v = append(v, vrx)
	}

	return v
}

func (g *Graph) TopologicalOrder() []*Vertex {
	return g.topologicalOrder
}

func (g *Graph) Clean() {
	g.topologicalOrder = nil
	g.vertices = nil
}

// AddVertex adds an vertex to the graph with its ID, value and meta information
func (g *Graph) AddVertex(vertex any, weight uint) {
	tp := reflect.TypeOf(vertex)
	g.vertices[tp] = &Vertex{
		id:     tp,
		value:  vertex,
		weight: weight,
		active: true,
	}
}

func (g *Graph) Remove(plugin any) []*Vertex {
	tp := reflect.TypeOf(plugin)
	var deletedVertices []*Vertex

	// remove the vertex from the graph
	vertex, ok := g.vertices[tp]
	if ok {
		delete(g.vertices, tp)
		deletedVertices = append(deletedVertices, vertex)
		vertex.active = false
	}

	edges := vertex.edges
	for i := range edges {
		if _, ok := g.vertices[reflect.TypeOf(edges[i].dest)]; !ok {
			continue
		}

		switch edges[i].connectionType {
		case InitConnection:
			p := edges[i].dest
			initMethod, _ := reflect.TypeOf(p).MethodByName(InitMethodName)

			args := make([]reflect.Type, initMethod.Type.NumIn())
			// receiver + other (should be other, since this is a dest vertex)
			for j := range initMethod.Type.NumIn() {
				args[j] = initMethod.Type.In(j)
			}

			// remove receiver
			args = args[1:]
		retry:
			for _, v := range g.vertices {
				if len(args) == 0 {
					break
				}
				if v.Plugin() == p || v.Plugin() == plugin {
					continue
				}

				for j := range args {
					if reflect.TypeOf(v.Plugin()).Implements(args[j]) {
						/*
							we've found a plugin which may replace our dependency
							now, since we modified the slice, start iteration again
						*/
						args = slices.Delete(args, j, j+1)
						goto retry
					}
				}
			}
			// we found replacement
			if len(args) == 0 {
				return deletedVertices
			}

			// we didn't find a replacement, mark the vertex as inactive
			deletedVertices = append(deletedVertices, g.vertices[reflect.TypeOf(edges[i].dest)])
			g.vertices[reflect.TypeOf(edges[i].dest)].active = false
			delete(g.vertices, reflect.TypeOf(edges[i].dest))
		case CollectsConnection:
			continue
		}
	}

	// remove all edges where dest is our plugin prepared to delete
	for _, v := range g.vertices {
		// since we are deleting while iterating, we need to create a new slice
		// to prevent index out of range or similar errors
		newEdges := make([]*edge, 0, len(v.edges))
		for i := range v.edges {
			if v.edges[i].dest == plugin {
				// we found an edge which is pointing to our plugin,
				// we need to remove it for the graph
				continue
			}
			newEdges = append(newEdges, v.edges[i])
		}
		v.edges = newEdges
	}

	for i := range g.topologicalOrder {
		for j := range deletedVertices {
			if g.topologicalOrder[i] == deletedVertices[j] {
				g.topologicalOrder[i].active = false
			}
		}
	}

	return deletedVertices
}

func (g *Graph) WriteDotString() {
	var s strings.Builder
	s.WriteString("digraph endure {\n")
	s.WriteString("\trankdir=TB;\n")
	s.WriteString("\tgraph [compound=true];\n")

	seenEdges := make(map[string]struct{})
	for i := range g.topologicalOrder {
		for j := range g.topologicalOrder[i].edges {
			src := reflect.TypeOf(g.topologicalOrder[i].edges[j].src).String()
			dest := reflect.TypeOf(g.topologicalOrder[i].edges[j].dest).String()

			if _, ok := seenEdges[src+dest]; !ok {
				s.WriteString(fmt.Sprintf("\t\"%s\" -> \"%s\";\n", src, dest))
				seenEdges[src+dest] = struct{}{}
			}
		}
	}
	s.WriteString("}\n")
	_, _ = fmt.Fprint(os.Stderr, s.String())
}
