package structures

import "reflect"

// manages the set of services and their edges
// type of the Graph: directed
type Graph struct {
	// nodes, which can have values
	// [a, b, c, etc..]5
	//graph    map[string]*Vertex
	Graph map[string]*Vertex
	// List of all Vertices
	Vertices []*Vertex

	// rows, connections
	// [a --> b], [a --> c] etc..
	// DEPENDENCIES
	Edges map[string][]string

	//graph    map[string]*Vertex

	// global property of the Graph
	// if the Graph Has disconnected nodes
	// this field will be set to true
	Connected bool
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
	RawPackage string
	// values to provide into INIT or Depends methods
	// key is a String() method invoked on the reflect.Value
	Values map[string]reflect.Value
}

// since we can have cyclic dependencies
// when we traverse the Graph, we should mark nodes as Visited or not to detect cycle
type Vertex struct {
	Id string
	// Value
	Value interface{}
	// Meta information about current Vertex
	Meta Meta
	// Dependencies of the node
	Dependencies []*Vertex
	// Visited used for the cyclic graphs to detect cycle
	Visited bool

	// Vertex foo4.S4 also provides (for example)
	// foo4.DB
	Provides map[string]*reflect.Value

	// for the toposort
	NumOfDeps int
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
		Graph:     make(map[string]*Vertex),
		Edges:     make(map[string][]string),
		Connected: false,
	}
}

func (g *Graph) HasVertex(name string) bool {
	_, ok := g.Graph[name]
	return ok
}

// BuildRunList builds run list from the graph after topological sort
// If Graph is not connected, separate lists could be run in parallel
func (g *Graph) BuildRunList() []*DoublyLinkedList {
	//graph := g.createServicesGraph()

	return nil
}


func (g *Graph) AddValue(vertexId, valueKey string, value reflect.Value) {
	// get the VERTEX
	if vertex, ok := g.Graph[vertexId]; ok {
		// add the vertex dep as value
		vertex.Meta.Values[valueKey] = value
	}
}


/*
AddDep doing the following:
1. Get a vertexID (foo2.S2 for example)
2. Get a depID --> could be vertexID of vertex dep ID like foo2.DB
3. Need to find VertexID to provide dependency. Example foo2.DB is actually foo2.S2 vertex
*/
func (g *Graph) AddDep(vertexID, depID string) {
	// idV should always present
	idV := g.GetVertex(vertexID)
	if idV == nil {
		panic("vertex should be in the graph")
	}
	// but depV can be represented like foo2.S2 (vertexID) or like foo2.DB (vertex foo2.S2, dependency foo2.DB)
	depV := g.GetVertex(depID)
	if depV == nil {
		depV = g.findVertexId(depID)
	}
	// append depID vertex
	for i := 0; i < len(idV.Dependencies); i++ {
		tmpId := idV.Dependencies[i].Id
		if tmpId == depV.Id {
			return
		}
	}
	idV.Dependencies = append(idV.Dependencies, depV)
	depV.NumOfDeps++
}

func (g *Graph) AddVertex(vertexId string, vertexValue interface{}, meta Meta) {
	g.Graph[vertexId] = &Vertex{
		// todo fill all the information
		Id:           vertexId,
		Value:        vertexValue,
		Meta:         meta,
		Dependencies: nil,
		Visited:      false,
	}
	g.Vertices = append(g.Vertices, g.Graph[vertexId])
}

func (g *Graph) GetVertex(id string) *Vertex {
	return g.Graph[id]
}

func (g *Graph) findVertexId(depId string) *Vertex {
	for i := 0; i < len(g.Vertices); i++ {
		//vertexId := g.Vertices[i].Id
		for id := range g.Vertices[i].Provides {
			if depId == id {
				return g.Vertices[i]
			}
		}
	}
	return nil
}

func (g *Graph) TopologicalSort() []string {
	var ord []string
	var verticesWoDeps []*Vertex

	for _, v := range g.Vertices {
		if v.NumOfDeps == 0 {
			verticesWoDeps = append(verticesWoDeps, v)
		}
	}

	for len(verticesWoDeps) > 0 {
		v := verticesWoDeps[len(verticesWoDeps)-1]
		verticesWoDeps = verticesWoDeps[:len(verticesWoDeps)-1]

		ord = append(ord, v.Id)
		g.removeDep(v, &verticesWoDeps)
	}

	return ord

}

func (g *Graph) removeDep(vertex *Vertex, verticesWoPrereqs *[]*Vertex) {
	for len(vertex.Dependencies) > 0 {
		dep := vertex.Dependencies[len(vertex.Dependencies)-1]
		vertex.Dependencies = vertex.Dependencies[:len(vertex.Dependencies)-1]
		dep.NumOfDeps--
		if dep.NumOfDeps == 0 {
			*verticesWoPrereqs = append(*verticesWoPrereqs, dep)
		}
	}
}
