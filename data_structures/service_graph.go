package data_structures

// manages the set of services and their edges
// type of the graph: directed
type graph struct {
	// nodes, which can have values
	nodes map[string]node
	// rows, connections
	edges map[string][]string

	// global property of the graph
	// if the graph has disconnected nodes
	// this field will be set to true
	connected bool
}

// since we can have cyclic dependencies
// when we traverse the graph, we should mark nodes os visited or not to detect cycle
type node struct {
	value   interface{}
	visited bool
}

func (g *graph) has(name string) bool {
	_, ok := g.nodes[name]
	return ok
}

func (g *graph) push(name string, node interface{}) {
	// todo temporary do not vidited
	g.nodes[name] = struct {
		value   interface{}
		visited bool
	}{value: node, visited: false}
	g.edges[name] = []string{}
}

func (g *graph) depends(name string, depends ...string) {
	for _, n := range depends {
		g.edges[name] = append(g.edges[name], n)
	}
}
