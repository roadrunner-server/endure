package cascade

// manages the set of services and their dependencies
type serviceGraph struct {
	nodes       map[string]interface{}
	dependecies map[string][]string
}

func (g *serviceGraph) has(name string) bool {
	_, ok := g.nodes[name]
	return ok
}

func (g *serviceGraph) push(name string, node interface{}) {
	g.nodes[name] = node
	g.dependecies[name] = []string{}
}

func (g *serviceGraph) depends(name string, depends ...string) {

}
