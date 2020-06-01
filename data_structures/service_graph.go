package data_structures

// manages the set of services and their edges
// type of the AdjacencyList: directed
type AdjacencyList struct {
	// nodes, which can have values
	// [a, b, c, etc..]
	Vertices map[string]Node
	// rows, connections
	// [a --> b], [a --> c] etc..
	Edges map[string][]string

	// global property of the AdjacencyList
	// if the AdjacencyList Has disconnected nodes
	// this field will be set to true
	Connected bool
}

// Meta information included into the Node
// May include:
// 1. Disabled info
// 2. Relation status
type Meta struct {
}

// since we can have cyclic dependencies
// when we traverse the AdjacencyList, we should mark nodes as Visited or not to detect cycle
type Node struct {
	// Value
	Value interface{}
	// Meta information about current Node
	Meta Meta
	// Visited used for the cyclic graphs to detect cycle
	Visited bool
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
func NewAL() *AdjacencyList {
	return &AdjacencyList{
		Vertices:  make(map[string]Node),
		Edges:     make(map[string][]string),
		Connected: false,
	}
}

func (g *AdjacencyList) Has(name string) bool {
	_, ok := g.Vertices[name]
	return ok
}

// tests whether there is an edge from the vertex x to the vertex y;
func (g *AdjacencyList) Adjacent() {

}

func (g *AdjacencyList) AddVertex(name string, node interface{}) {
	// todo temporary do not visited
	g.Vertices[name] = struct {
		Value   interface{}
		Meta    Meta
		Visited bool
	}{
		Value:   node,
		Visited: false,
		Meta:    Meta{},
	}
	g.Edges[name] = []string{}
}

func (g *AdjacencyList) AddEdge(name string, depends ...string) {
	for _, n := range depends {
		g.Edges[name] = append(g.Edges[name], n)
	}
}

// Find will return pointer to the Node or nil, if the Node does not exist
// O(V+E) time complexity
// O(V) space complexity
//func (g *AdjacencyList) FindDFS(name string) *Node {
//	for k, v := range g.Vertices {
//		if k == name {
//
//		} else {
//			g.FindDFS(name)
//		}
//	}
//}

// BuildRunList builds run list from the graph after topological sort
// If AdjacencyList is not connected, separate lists could be run in parallel
func (g *AdjacencyList) BuildRunList() []*DoublyLinkedList {
	return nil
}
