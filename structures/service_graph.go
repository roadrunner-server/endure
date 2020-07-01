package structures

import (
	"fmt"
	"reflect"
)

type Kind int

const (
	Init Kind = iota
	Depends
)

// manages the set of services and their edges
// type of the Graph: directed
type Graph struct {
	// nodes, which can have values
	// [a, b, c, etc..]5
	//graph    map[string]*Vertex
	Graph map[string]*Vertex
	// List of all Vertices
	Vertices []*Vertex
}

// it results in "RPC" --> S1, and at the end slice with deps will looks like:
// []deps{Dep{"RPC", S1}, Dep{"RPC", S2"}..etc}
// SHOULD BE IN GRAPH
//type Dep struct {
//	Id string      // for example rpc
//	D  interface{} // S1
//}

// Meta information included into the Vertex
// May include:
// 1. Disabled info
// 2. Relation status
type Meta struct {
	// FnsProviderToInvoke is the function names to invoke if type implements Provides() interface
	FnsProviderToInvoke []string
	// FnsRegisterToInvoke is the function names to invoke if type implements Register() interface
	FnsRegisterToInvoke []string

	// List of the vertex deps
	// foo4.DB, foo4.S4 etc.. which were found in the Init() method
	InitDepsList []DepsEntry

	// List of the vertex deps
	// foo4.DB, foo4.S4 etc.. which were found in the Depends() method
	DepsList []DepsEntry
}

type DepsEntry struct {
	Name        string
	IsReference *bool
}

// since we can have cyclic dependencies
// when we traverse the Graph, we should mark nodes as Visited or not to detect cycle
type Vertex struct {
	Id string
	// Vertex
	Iface interface{}
	// Meta information about current Vertex
	Meta Meta
	// Dependencies of the node
	Dependencies []*Vertex

	// Vertex foo4.S4 also provides (for example)
	// foo4.DB
	Provides map[string]ProvidedEntry

	// for the toposort
	NumOfDeps int

	Visited  bool
	Visiting bool
}

type ProvidedEntry struct {
	// we need to distinguish false (default bool value) and nil --> we don't know information about reference
	IsReference *bool
	Value       *reflect.Value
}

func (v *Vertex) AddValue(valueKey string, value reflect.Value, isRef bool) error {
	// get the VERTEX
	if v.Provides == nil {
		v.Provides = make(map[string]ProvidedEntry)
	}

	v.Provides[valueKey] = ProvidedEntry{
		IsReference: &isRef,
		Value:       &value,
	}
	return nil
}

// NewAL initializes adjacency list to store the Graph
// example
// 1 -> 2 -> 4
// 2 -> 5
// 3 -> 6 -> 5
// 4 -> 2
// 5 -> 4
// 6 -> 6
//
// Graph from the AL:
//
//+---+          +---+               +---+
//| 1 +--------->+ 2 |               | 3 |
//+-+-+          +--++               +-+-+
//  |          +-+  |             +-+  |
//  |        +-+    |           +-+    |
//  |       ++      |          ++      |
//  v     +-+       v        +-+       v
//+-+-+<--+      +--++       |       +-+-+
//| 4 |     +----+ 5 +<------+       | 6 +<-+
//+---+<----+    +---+               +-+-+  |
//                                     |    |
//                                     +----+
// BUT
// According to the topological sorting, graph should be
// 1. DIRECTED
// 2. ACYCLIC
//
func NewGraph() *Graph {
	return &Graph{
		Graph: make(map[string]*Vertex),
	}
}

func (g *Graph) HasVertex(name string) bool {
	_, ok := g.Graph[name]
	return ok
}

/*
AddDepRev doing the following:
1. Get a vertexID (foo2.S2 for example)
2. Get a depID --> could be vertexID of vertex dep ID like foo2.DB
3. Need to find VertexID to provide dependency. Example foo2.DB is actually foo2.S2 vertex
*/
func (g *Graph) AddDep(vertexID, depID string, kind Kind, isRef bool) error {
	// vertex should always present
	vertex := g.GetVertex(vertexID)
	if vertex == nil {
		panic("vertex should be in the graph")
	}
	// but depVertex can be represented like foo2.S2 (vertexID) or like foo2.DB (vertex foo2.S2, dependency foo2.DB)
	depVertex := g.GetVertex(depID)
	if depVertex == nil {
		depVertex = g.FindProvider(depID)
	}
	if depVertex == nil {
		return fmt.Errorf("can't find dep: %s for the vertex: %s", depID, vertexID)
	}

	// add Dependency into the List
	// to call later
	// because we should know Init method parameters for every Vertex
	switch kind {
	case Init:
		if vertex.Meta.InitDepsList == nil {
			vertex.Meta.InitDepsList = make([]DepsEntry, 0, 1)
		}
		vertex.Meta.InitDepsList = append(vertex.Meta.InitDepsList, DepsEntry{
			Name:        depID,
			IsReference: &isRef,
		})
	case Depends:
		if vertex.Meta.DepsList == nil {
			vertex.Meta.DepsList = make([]DepsEntry, 0, 1)
		}
		vertex.Meta.DepsList = append(vertex.Meta.DepsList, DepsEntry{
			Name:        depID,
			IsReference: &isRef,
		})
	}

	// append depID vertex
	for i := 0; i < len(depVertex.Dependencies); i++ {
		tmpId := depVertex.Dependencies[i].Id
		if tmpId == vertex.Id {
			return nil
		}
	}
	//depVertex.NumOfDeps++
	//vertex.Dependencies = append(vertex.Dependencies, depVertex)
	vertex.NumOfDeps++
	depVertex.Dependencies = append(depVertex.Dependencies, vertex)
	return nil
}

func (g *Graph) AddDepRev(vertexID, depID string, kind Kind, isRef bool) error {
	// idV should always present
	idV := g.GetVertex(vertexID)
	if idV == nil {
		panic("vertex should be in the graph")
	}
	// but depV can be represented like foo2.S2 (vertexID) or like foo2.DB (vertex foo2.S2, dependency foo2.DB)
	depV := g.GetVertex(depID)
	if depV == nil {
		depV = g.FindProvider(depID)
	}
	if depV == nil {
		return fmt.Errorf("can't find dep: %s for the vertex: %s", depID, vertexID)
	}

	// add Dependency into the List
	// to call later
	// because we should know Init method parameters for every Vertex
	switch kind {
	case Init:
		if idV.Meta.InitDepsList == nil {
			idV.Meta.InitDepsList = make([]DepsEntry, 0, 1)
		}
		idV.Meta.InitDepsList = append(idV.Meta.InitDepsList, DepsEntry{
			Name:        depID,
			IsReference: &isRef,
		})
	case Depends:
		if idV.Meta.DepsList == nil {
			idV.Meta.DepsList = make([]DepsEntry, 0, 1)
		}
		idV.Meta.DepsList = append(idV.Meta.DepsList, DepsEntry{
			Name:        depID,
			IsReference: &isRef,
		})
	}

	// append depID vertex
	for i := 0; i < len(idV.Dependencies); i++ {
		tmpId := idV.Dependencies[i].Id
		if tmpId == depV.Id {
			return nil
		}
	}
	depV.NumOfDeps++
	idV.Dependencies = append(idV.Dependencies, depV)
	return nil
}

func (g *Graph) AddVertex(vertexId string, vertexIface interface{}, meta Meta) {
	g.Graph[vertexId] = &Vertex{
		Id:           vertexId,
		Iface:        vertexIface,
		Meta:         meta,
		Dependencies: nil,
	}
	g.Vertices = append(g.Vertices, g.Graph[vertexId])
}

func (g *Graph) GetVertex(id string) *Vertex {
	return g.Graph[id]
}

func (g *Graph) FindProvider(depId string) *Vertex {
	for i := 0; i < len(g.Vertices); i++ {
		for providerId := range g.Vertices[i].Provides {
			if depId == providerId {
				return g.Vertices[i]
			}
		}
	}
	return nil
}

// deprecated
func OldTopologicalSort(vertices []*Vertex) []*Vertex {
	var ord []*Vertex
	var verticesWoDeps []*Vertex

	for _, v := range vertices {
		if v.NumOfDeps == 0 {
			verticesWoDeps = append(verticesWoDeps, v)
		}
	}

	for len(verticesWoDeps) > 0 {
		v := verticesWoDeps[len(verticesWoDeps)-1]
		verticesWoDeps = verticesWoDeps[:len(verticesWoDeps)-1]

		ord = append(ord, v)
		removeDep(v, &verticesWoDeps)
	}

	return ord

}

func TopologicalSort(vertices []*Vertex) []*Vertex {
	var ord []*Vertex
	verticesCopy := vertices

	for len(verticesCopy) != 0 {
		vertex := verticesCopy[len(verticesCopy)-1]
		verticesCopy = verticesCopy[:len(verticesCopy)-1]
		containsCycle := dfs(vertex, &ord)
		if containsCycle {
			return nil
		}
	}

	return ord
}

func dfs(vertex *Vertex, ordered *[]*Vertex) bool {
	if vertex.Visited {
		return false
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

func removeDep(vertex *Vertex, verticesWoPrereqs *[]*Vertex) {
	for i := 0; i < len(vertex.Dependencies); i++ {
		dep := vertex.Dependencies[i]
		dep.NumOfDeps--
		if dep.NumOfDeps == 0 {
			*verticesWoPrereqs = append(*verticesWoPrereqs, dep)
		}
	}
	// TODO remove dependencies thus we don't need it in the run list
	//for len(vertex.Dependencies) > 0 {
	//	dep := vertex.Dependencies[len(vertex.Dependencies)-1]
	//	//vertex.Dependencies = vertex.Dependencies[:len(vertex.Dependencies)-1]
	//	dep.NumOfDeps--
	//	if dep.NumOfDeps == 0 {
	//		*verticesWoPrereqs = append(*verticesWoPrereqs, dep)
	//	}
	//}
}
