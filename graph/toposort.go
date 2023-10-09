package graph

func (g *Graph) TopologicalSort() {
	heap := &VertexHeap{}

	for _, v := range g.vertices {
		if v.indegree == 0 {
			heap.Push(v)
		}
	}

	processed := make(map[*Vertex]struct{}, 2)
	for heap.Len() > 0 {
		// Remove the vertex with the smallest number of edges and the highest edge weight from the priority queue.
		v, _ := heap.Pop().(*Vertex)

		if _, ok := processed[v]; ok {
			continue
		}

		// Add the vertex to the topological order.
		g.topologicalOrder = append(g.topologicalOrder, v)
		processed[v] = struct{}{}

		// Decrement the indegree of each of the vertex's neighbors.
		// If a neighbor's indegree becomes 0, add it to the priority queue.
		for i := 0; i < len(v.edges); i++ {
			dest := g.VertexById(v.edges[i].dest)
			dest.indegree--

			if dest.indegree == 0 {
				heap.Push(dest)
			}
		}
	}
}
