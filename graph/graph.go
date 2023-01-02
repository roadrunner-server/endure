package graph

import (
	"reflect"
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
	var v []*Vertex

	for _, vrx := range g.vertices {
		v = append(v, vrx)
	}

	return v
}

func (g *Graph) TopologicalOrder() []*Vertex {
	return g.topologicalOrder
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
	for i := 0; i < len(edges); i++ {
		if _, ok := g.vertices[reflect.TypeOf(edges[i].dest)]; !ok {
			continue
		}

		switch edges[i].connectionType {
		case InitConnection:
			deletedVertices = append(deletedVertices, g.vertices[reflect.TypeOf(edges[i].dest)])
			g.vertices[reflect.TypeOf(edges[i].dest)].active = false
			delete(g.vertices, reflect.TypeOf(edges[i].dest))
		case CollectsConnection:
			continue
		}
	}

	for i := 0; i < len(g.topologicalOrder); i++ {
		for j := 0; j < len(deletedVertices); j++ {
			if g.topologicalOrder[i] == deletedVertices[j] {
				g.topologicalOrder[i].active = false
			}
		}
	}

	return deletedVertices
}
