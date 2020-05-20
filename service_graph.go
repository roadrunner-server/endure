package cascade

// manages the set of services and their dependencies
type serviceGraph struct {
	nodes        map[string]interface{}
	dependencies map[string][]string
}

func (g *serviceGraph) has(name string) bool {
	_, ok := g.nodes[name]
	return ok
}

func (g *serviceGraph) push(name string, node interface{}) {
	g.nodes[name] = node
	g.dependencies[name] = []string{}
}

func (g *serviceGraph) depends(name string, depends ...string) {
	for _, n := range depends {
		g.dependencies[name] = append(g.dependencies[name], n)
	}
}
