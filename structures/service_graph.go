package structures

// manages the set of services and their edges
// type of the Graph: directed
type Graph struct {
	// nodes, which can have values
	// [a, b, c, etc..]
	Vertices map[string]Vertex
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

// it results in "RPC" --> S1, and at the end slice with deps will looks like:
// []deps{Dep{"RPC", S1}, Dep{"RPC", S2"}..etc}
type Dep struct {
	Id string      // for example rpc
	D  interface{} // S1
}

func NewDeps() Dep {
	return Dep{}
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
		Vertices:  make(map[string]Vertex),
		Edges:     make(map[string][]string),
		Connected: false,
	}
}

func (g *Graph) Has(name string) bool {
	_, ok := g.Vertices[name]
	return ok
}

// tests whether there is an edge from the vertex x to the vertex y;
func (g *Graph) Adjacent() {

}

func (g *Graph) AddVertex(name string, value interface{}, raw string) {
	// todo temporary do not visited
	g.Vertices[name] = struct {
		Id           string
		Value        interface{}
		Meta         Meta
		Dependencies []*Vertex
		Visited      bool
		NumOfDeps    int
	}{
		Value:   value,
		Visited: false,
		Meta: Meta{
			RawPackage: raw,
		},
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
