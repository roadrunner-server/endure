package data_structures

// manages the set of services and their edges
// type of the Graph: directed
type Graph struct {
	// nodes, which can have values
	Nodes map[string]Node
	// rows, connections
	Edges map[string][]string

	// global property of the Graph
	// if the Graph Has disconnected nodes
	// this field will be set to true
	Connected bool
}

// since we can have cyclic dependencies
// when we traverse the Graph, we should mark nodes as Visited or not to detect cycle
type Node struct {
	Value   interface{}
	Visited bool
}

func (g *Graph) Has(name string) bool {
	_, ok := g.Nodes[name]
	return ok
}

func (g *Graph) Push(name string, node interface{}) {
	// todo temporary do not vidited
	g.Nodes[name] = struct {
		Value   interface{}
		Visited bool
	}{Value: node, Visited: false}
	g.Edges[name] = []string{}
}

func (g *Graph) Depends(name string, depends ...string) {
	for _, n := range depends {
		g.Edges[name] = append(g.Edges[name], n)
	}
}
