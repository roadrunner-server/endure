package graph

import (
	"reflect"
)

// Vertex is the main vertex representation for the graph
// since we can have cyclic dependencies
// when we traverse the VerticesMap, we should mark nodes as visited or not to detect cycle
type Vertex struct {
	// ID of the vertex, currently string representation of the structure fn
	id reflect.Type

	// value is a plugin itself
	value any
	// edges
	edges []*edge
	// for the topological sort, private
	// https://en.wikipedia.org/wiki/Directed_graph#Indegree_and_outdegree
	indegree int
	// vertex-weighted graph
	weight uint
	// active represents the current state of the vertex
	active bool
}

func (v *Vertex) ID() reflect.Type {
	return v.id
}

func (v *Vertex) Plugin() any {
	return v.value
}

func (v *Vertex) Weight() uint {
	return v.weight
}

func (v *Vertex) IsActive() bool {
	return v.active
}
