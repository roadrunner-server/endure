package structures

// manages the set of services and their edges
// type of the Graph: directed
type Graph struct {
	// nodes, which can have values
	// [a, b, c, etc..]
	Vertices map[string]*Vertex
	// rows, connections
	// [a --> b], [a --> c] etc..
	Edges map[string][]string

	// global property of the Graph
	// if the Graph Has disconnected nodes
	// this field will be set to true
	Connected bool
}

// Meta information included into the Vertex
// May include:
// 1. Disabled info
// 2. Relation status
type Meta struct {
	RawPackage string
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
func NewAL() *Graph {
	return &Graph{
		Vertices:  make(map[string]*Vertex),
		Edges:     make(map[string][]string),
		Connected: false,
	}
}

func (g *Graph) Has(name string) bool {
	_, ok := g.Vertices[name]
	return ok
}

func (g *Graph) AddVertex(name string, value interface{}, meta Meta) {
	// todo temporary do not visited
	g.Vertices[name] = &Vertex{
		Id:           "",
		Value:        value,
		Meta:         meta,
		Dependencies: nil,
		Visited:      false,
		NumOfDeps:    0,
	}
	// initialization
	g.Edges[name] = []string{}
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

type depsGraph struct {
	vertices []*Vertex
	graph    map[string]*Vertex
}

// it results in "RPC" --> S1, and at the end slice with deps will looks like:
// []deps{Dep{"RPC", S1}, Dep{"RPC", S2"}..etc}
type Dep struct {
	Id string      // for example rpc
	D  interface{} // S1
}

// []string here all the deps (vertices) S1, S2, S3, S4
func NewDepsGraph(deps []string) *depsGraph {
	g := &depsGraph{
		vertices: make([]*Vertex, 0, 10),
		graph:    make(map[string]*Vertex),
	}
	for _, d := range deps {
		g.AddVertex(d)
	}
	return g
}

func (g *depsGraph) AddDep(id, dep string) {
	// get vertex for ID and for the deps
	idV, depV := g.GetVertex(id), g.GetVertex(dep)
	// append dep vertex
	idV.Dependencies = append(idV.Dependencies, depV)
	depV.NumOfDeps++
}

func (g *depsGraph) AddVertex(id string) {
	g.graph[id] = &Vertex{
		// todo fill all the information
		Id:           id,
		Value:        nil,
		Meta:         Meta{},
		Dependencies: nil,
		Visited:      false,
	}
	g.vertices = append(g.vertices, g.graph[id])
}

func (g *depsGraph) GetVertex(id string) *Vertex {
	if _, found := g.graph[id]; !found {
		g.AddVertex(id)
	}

	return g.graph[id]
}

func (g *depsGraph) Order() []string {
	var ord []string
	var verticesWoDeps []*Vertex

	for _ ,v := range g.vertices {
		if v.NumOfDeps == 0 {
			verticesWoDeps = append(verticesWoDeps, v)
		}
	}

	for len(verticesWoDeps) > 0 {
		v := verticesWoDeps[len(verticesWoDeps) - 1]
		verticesWoDeps = verticesWoDeps[:len(verticesWoDeps) - 1]

		ord = append(ord, v.Id)
		g.removeDep(v, &verticesWoDeps)
	}

	return ord

}

func (g *depsGraph) removeDep(vertex *Vertex, verticesWoPrereqs *[]*Vertex) {
	for len(vertex.Dependencies) > 0 {
		dep := vertex.Dependencies[len(vertex.Dependencies) - 1]
		vertex.Dependencies = vertex.Dependencies[:len(vertex.Dependencies) - 1]
		dep.NumOfDeps--
		if dep.NumOfDeps == 0 {
			*verticesWoPrereqs = append(*verticesWoPrereqs, dep)
		}
	}
}