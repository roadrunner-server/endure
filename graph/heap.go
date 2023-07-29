package graph

import (
	"container/heap"
)

type VertexHeap []*Vertex

func (h *VertexHeap) Len() int {
	return len(*h)
}
func (h *VertexHeap) Less(i, j int) bool {
	return (*h)[i].weight < (*h)[j].weight
}

func (h *VertexHeap) Swap(i, j int) {
	(*h)[i], (*h)[j] = (*h)[j], (*h)[i]
}

func (h *VertexHeap) Push(x interface{}) {
	*h = append(*h, x.(*Vertex))
	heap.Fix(h, len(*h)-1)
}

func (h *VertexHeap) Pop() interface{} {
	n := len(*h)
	x := (*h)[n-1]
	*h = (*h)[0 : n-1]
	return x
}
