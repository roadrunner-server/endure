package data_structures

// manages the set of services and their edges
// type of the Graph: directed
type Graph struct {
	// nodes, which can have values
	// [a, b, c, etc..]
	Vertices map[string]Node
	// rows, connections
	// [a --> b], [a --> c] etc..
	Edges map[string][]string

	// global property of the Graph
	// if the Graph Has disconnected nodes
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
// when we traverse the Graph, we should mark nodes as Visited or not to detect cycle
type Node struct {
	// Value
	Value   interface{}
	// Meta information about current Node
	Meta    Meta
	// Visited used for the cyclic graphs to detect cycle
	Visited bool
}

func NewGraph() *Graph {
	return &Graph{
		Vertices:  nil,
		Edges:     nil,
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

func (g *Graph) AddVertex(name string, node interface{}) {
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

func (g *Graph) AddEdge(name string, depends ...string) {
	for _, n := range depends {
		g.Edges[name] = append(g.Edges[name], n)
	}
}

// Find will return pointer to the Node or nil, if the Node does not exist
// O(V+E) time complexity
// O(V) space complexity
func (g *Graph) FindDFS(name string) *Node {
	for k, v := range g.Vertices {
		if k == name {

		} else {
			g.FindDFS(name)
		}
	}
}

// BuildRunList builds run list from the graph
// If Graph is not connected, separate lists could be run in parallel
func (g *Graph) BuildRunList() []*DoublyLinkedList {
	return nil
}
