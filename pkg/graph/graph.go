package graph

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/spiral/endure/pkg/fsm"
	"github.com/spiral/endure/pkg/vertex"
	"github.com/spiral/errors"
)

// Kind used to identify the method which invokes AddDeps
type Kind int

const (
	// Init method
	Init Kind = iota
	// Collects method
	Collects
)

// Graph manages the set of services and their edges
// type of the VerticesMap: directed
type Graph struct {
	// Map with vertices to have an easy access to it
	VerticesMap map[string]*vertex.Vertex
	// List of all Vertices
	Vertices []*vertex.Vertex

	Providers map[string]reflect.Value

	// internal type
	InitialProviders map[string]reflect.Value
}

// NewGraph initializes endure Graph
// According to the topological sorting, graph should be
// 1. DIRECTED
// 2. ACYCLIC
func NewGraph() *Graph {
	return &Graph{
		VerticesMap: make(map[string]*vertex.Vertex),
		Providers:   make(map[string]reflect.Value),
		Vertices:    make([]*vertex.Vertex, 0),
	}
}

// AddGlobalProvider adds provider to the global map in the Graph structure
func (g *Graph) AddGlobalProvider(providedID string, val reflect.Value) {
	g.Providers[providedID] = val
}

// HasVertex returns true or false if the vertex exists in the vertices map in the graph
func (g *Graph) HasVertex(name string) bool {
	_, ok := g.VerticesMap[name]
	return ok
}

func (g *Graph) ClearState() {
	g.Providers = make(map[string]reflect.Value)
	g.VerticesMap = make(map[string]*vertex.Vertex)
	// delete from Vertices
	g.Vertices = make([]*vertex.Vertex, 0)
}

/*
AddInterfaceDep doing the following:
1. Get a vertexID (foo2.S2 for example)
2. Get a depID --> could be vertexID of vertex dep ID like foo2.DB
3. Need to find vertexID to provide dependency. Example foo2.DB is actually foo2.S2 vertex
*/
func (g *Graph) AddInterfaceDep(vertex *vertex.Vertex, depID string, method Kind, isRef bool) error {
	const op = errors.Op("endure_add_interface_dep")
	// here can be a lot of deps
	depVertex := g.FindProviders(depID)
	if depVertex == nil {
		return errors.E(op, errors.Errorf("can't find dependency: %s for the vertex: %s", depID, vertex.ID))
	}

	// skip self
	if depVertex.ID == vertex.ID {
		return nil
	}

	// add Dependency into the List
	// to call later
	// because we should know Init method parameters for every Vertex
	// for example, we should know http.Middleware dependency and later invoke all types which it implement
	// OR know Collects methods to invoke
	if !g.addToList(method, vertex, depID, isRef, depVertex.ID, reflect.Interface) {
		return nil
	}

	vertex.NumOfDeps++
	vertex.Dependencies = append(vertex.Dependencies, depVertex)
	return nil
}

func (g *Graph) AddStructureDep(vertex *vertex.Vertex, depID string, method Kind, isRef bool) error {
	const op = errors.Op("endure_add_structure_dep")
	// vertex should always present

	// but depVertex can be represented like foo2.S2 (vertexID) or like foo2.DB (vertex foo2.S2, dependency foo2.DB)
	depVertex := g.GetVertex(depID)
	if depVertex == nil {
		tmp := g.FindProviders(depID)
		if tmp == nil {
			return errors.E(op, errors.Errorf("can't find dep: %s for the vertex: %s", depID, vertex.ID))
		}
		depVertex = tmp
	}

	// add Dependency into the List
	// to call later
	// because we should know Init method parameters for every Vertex
	if !g.addToList(method, vertex, depID, isRef, depVertex.ID, reflect.Struct) {
		return nil
	}

	vertex.NumOfDeps++
	vertex.Dependencies = append(vertex.Dependencies, depVertex)
	return nil
}

// Add meta information to the InitDepsToInvoke or CollectsDepsToInvoke
func (g *Graph) addToList(method Kind, vrtx *vertex.Vertex, depID string, isRef bool, refID string, kind reflect.Kind) bool {
	switch method {
	case Init:
		addInit(vrtx, refID, depID, isRef, kind)
	case Collects:
		if _, ok := vrtx.Meta.CollectsDepsToInvoke[refID]; ok {
			return false
		}
		vrtx.Meta.CollectsDepsToInvoke[refID] = append(vrtx.Meta.CollectsDepsToInvoke[refID], vertex.Entry{
			RefID:       refID,
			Name:        depID,
			IsReference: &isRef,
			Kind:        kind,
		})
		contains := false
		for _, v := range vrtx.Meta.InitDepsOrd {
			if v == refID {
				contains = true
			}
		}
		if !contains {
			vrtx.Meta.InitDepsOrd = append(vrtx.Meta.InitDepsOrd, refID)
		}
	}
	return true
}

func addInit(vrtx *vertex.Vertex, refID string, depID string, isRef bool, kind reflect.Kind) {
	vrtx.Meta.InitDepsToInvoke[refID] = append(vrtx.Meta.InitDepsToInvoke[refID], vertex.Entry{
		RefID:       refID,
		Name:        depID,
		IsReference: &isRef,
		Kind:        kind,
	})
	contains := false
	for _, v := range vrtx.Meta.InitDepsOrd {
		if v == refID {
			contains = true
		}
	}
	if !contains {
		vrtx.Meta.InitDepsOrd = append(vrtx.Meta.InitDepsOrd, refID)
	}
}

// Reset resets vertices to initial state
func (g *Graph) Reset(vrtx *vertex.Vertex) []*vertex.Vertex {
	// restore number of dependencies for the root
	vrtx.NumOfDeps = len(vrtx.Dependencies)
	vrtx.Visiting = false
	vrtx.Visited = false
	vrtx.SetState(fsm.Uninitialized)
	vertices := make([]*vertex.Vertex, 0, 5)
	vertices = append(vertices, vrtx)

	tmp := make(map[string]*vertex.Vertex)

	g.depthFirstSearch(vrtx.Dependencies, tmp)

	for _, v := range tmp {
		vertices = append(vertices, v)
	}
	return vertices
}

// actually this is DFS just to reset all vertices to initial state after topological sort
func (g *Graph) depthFirstSearch(deps []*vertex.Vertex, tmp map[string]*vertex.Vertex) {
	for i := 0; i < len(deps); i++ {
		deps[i].Visited = false
		deps[i].Visiting = false
		deps[i].NumOfDeps = len(deps)
		tmp[deps[i].ID] = deps[i]
		g.depthFirstSearch(deps[i].Dependencies, tmp)
	}
}

// AddVertex adds an vertex to the graph with its ID, value and meta information
func (g *Graph) AddVertex(vertexID string, vertexIface interface{}) {
	v := vertex.NewVertex()
	v.ID = vertexID
	v.Iface = vertexIface
	v.SetState(fsm.Uninitialized)

	m := vertex.NewMeta()
	v.Meta = m

	g.VerticesMap[vertexID] = v
	g.Vertices = append(g.Vertices, g.VerticesMap[vertexID])
}

// GetVertex returns vertex by its ID
func (g *Graph) GetVertex(id string) *vertex.Vertex {
	return g.VerticesMap[id]
}

// FindProviders finds provider deps for the vertex and returns dependent vertices
func (g *Graph) FindProviders(depID string) *vertex.Vertex {
	for i := 0; i < len(g.Vertices); i++ {
		for providerID := range g.Vertices[i].Provides {
			if depID == providerID {
				return g.Vertices[i]
			}
		}
	}

	// try to find directly in the graph
	if _, ok := g.VerticesMap[depID]; ok {
		return g.VerticesMap[depID]
	}

	return nil
}

// TopologicalSort topologically sort the graph and return slice of the sorted vertices
func TopologicalSort(vertices []*vertex.Vertex) ([]*vertex.Vertex, error) {
	const op = errors.Op("graph_topological_sort")
	var ord []*vertex.Vertex
	verticesCopy := vertices

	buf := new(strings.Builder)
	defer buf.Reset()

	for len(verticesCopy) != 0 {
		vrtx := verticesCopy[len(verticesCopy)-1]
		verticesCopy = verticesCopy[:len(verticesCopy)-1]
		containsCycle := dfs(vrtx, &ord)
		if containsCycle {
			// allocate a buffer for the resulting message
			// defer buffer reset
			buf.WriteString("The following vertices involved:\n")
			// If we found a cycle, print involved vertices in reverse order
			for i := len(vertices) - 1; i > 0; i-- {
				if !vertices[i].Visited {
					buf.WriteString(fmt.Sprintf("vertex: %s -> ", vertices[i].ID))
				}
			}

			// trim the last arrow and return error message
			return nil, errors.E(op, errors.Errorf("cycle detected, please, check the path: %s", strings.TrimRight(buf.String(), "-> ")))
		}
	}

	return ord, nil
}

func dfs(vertex *vertex.Vertex, ordered *[]*vertex.Vertex) bool {
	// if vertex already visited, we reach the end, return
	if vertex.Visited {
		return false
		// if vertex is visiting, than we have cyclic dep, because we reach the same vertex from two different dependencies
	} else if vertex.Visiting {
		return true
	}
	vertex.Visiting = true
	for _, depV := range vertex.Dependencies {
		containsCycle := dfs(depV, ordered)
		if containsCycle {
			return true
		}
	}
	vertex.Visited = true
	vertex.Visiting = false
	*ordered = append(*ordered, vertex)
	return false
}
