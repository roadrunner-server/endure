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
type Dep struct {
	Id string      // for example rpc
	D  interface{} // S1
}

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
	Provides []string

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

func (g *Graph) Has(name string) bool {
	_, ok := g.Graph[name]
	return ok
}

func (g *Graph) AddEdge(name string, depends ...string) {
	for _, n := range depends {
		g.Edges[name] = append(g.Edges[name], n)
	}
}

// BuildRunList builds run list from the graph after topological sort
// If Graph is not connected, separate lists could be run in parallel
func (g *Graph) BuildRunList() []*DoublyLinkedList {
	//graph := g.createServicesGraph()

	return nil
}

// []string here all the deps (vertices) S1, S2, S3, S4
//func NewDepsGraph(deps []string) *depsGraph {
//	g := &depsGraph{
//		vertices: make([]*Vertex, 0, 10),
//		graph:    make(map[string]*Vertex),
//	}
//	for _, d := range deps {
//		g.AddVertex(d)
//	}
//	return g
//}

func (g *Graph) AddValue(vertexId, valueKey string, value reflect.Value) {
	// get the VERTEX
	if vertex, ok := g.Graph[vertexId]; ok {
		// add the vertex dep as value
		vertex.Meta.Values[valueKey] = value
	}
}

//func (g *Graph) Graph() []*Vertex {
//	return g.vertices
//}

func (g *Graph) AddDep(vertexId, vertexDepId string) {
	// get vertex for ID and for the deps
	idV, depV := g.GetVertex(vertexId), g.GetVertex(vertexDepId)
	// append vertexDepId vertex
	idV.Dependencies = append(idV.Dependencies, depV)
	depV.NumOfDeps++
}

func (g *Graph) AddVertex(vertexId string, vertexValue interface{}, meta Meta) {

	//// todo temporary do not visited
	//g.Graph[name] = &Vertex{
	//	Id:           "",
	//	Value:        value,
	//	Meta:         meta,
	//	Dependencies: nil,
	//	Visited:      false,
	//	NumOfDeps:    0,
	//}
	//// initialization
	//g.Edges[name] = []string{}

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
	if g.Graph[id] == nil {
		g.Graph[id] = &Vertex{}
		//g.AddVertex(id)
	}
	//if _, found := g.Graph[id]; !found {
	//	g.AddVertex(id)
	//}

	return g.Graph[id]
}

func (g *Graph) Order() []string {
	var ord []string
	//var verticesWoDeps []*Vertex

	//for _, v := range g.vertices {
	//	if v.NumOfDeps == 0 {
	//		verticesWoDeps = append(verticesWoDeps, v)
	//	}
	//}
	//
	//for len(verticesWoDeps) > 0 {
	//	v := verticesWoDeps[len(verticesWoDeps)-1]
	//	verticesWoDeps = verticesWoDeps[:len(verticesWoDeps)-1]
	//
	//	ord = append(ord, v.Id)
	//	g.removeDep(v, &verticesWoDeps)
	//}

	return ord

}

//func (c *Cascade) topologicalSort() []string {
//	ids := make([]string, 0)
//	for k, _ := range c.servicesGraph.Graph {
//		ids = append(ids, k)
//	}
//
//	gr := structures.NewDepsGraph(ids)
//
//	for id, dep := range c.deps {
//		for _, v := range dep {
//			if v.D == nil {
//				continue
//			}
//			gr.AddDep(id, v.D.(string))
//		}
//	}
//
//	c.depsGraph = gr.Graph()
//
//	return gr.Order()
//}

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
