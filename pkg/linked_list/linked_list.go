package linked_list

import "github.com/spiral/endure/pkg/vertex"

// DllNode consists of the curr Vertex, Prev and Next DllNodes
type DllNode struct {
	Vertex     *vertex.Vertex
	Prev, Next *DllNode
}

// DoublyLinkedList is the node of DLL which is connected to the tail and the head
type DoublyLinkedList struct {
	Head, Tail *DllNode
}

// NewDoublyLinkedList returns DLL implementation
func NewDoublyLinkedList() *DoublyLinkedList {
	return &DoublyLinkedList{}
}

// SetHead O(1) time + space
func (dll *DoublyLinkedList) SetHead(node *DllNode) {
	if dll.Head == nil {
		dll.Head = node
		dll.Tail = node
		return
	}
	dll.InsertBefore(dll.Head, node)
}

// Push used to push vertex to the head
func (dll *DoublyLinkedList) Push(vertex *vertex.Vertex) {
	node := &DllNode{
		Vertex: vertex,
	}

	if dll.Head == nil {
		dll.Head = node
		dll.Tail = node
		return
	}

	node.Next = dll.Head
	dll.Head.Prev = node

	node.Prev = nil

	dll.Head = node
}

// PushTail used to push vertex to the tail
func (dll *DoublyLinkedList) PushTail(vertex *vertex.Vertex) {
	node := &DllNode{
		Vertex: vertex,
	}
	node.Next = dll.Tail
	dll.Tail.Next = node

	node.Prev = dll.Head

	dll.Tail = node
}

// SetTail sets the tail, constant O(1) time and space
func (dll *DoublyLinkedList) SetTail(node *DllNode) {
	if dll.Tail == nil {
		dll.SetHead(node)
	}
	dll.InsertAfter(dll.Tail, node)
}

// InsertBefore inserts node before the provided node
func (dll *DoublyLinkedList) InsertBefore(node, nodeToInsert *DllNode) {
	if nodeToInsert == dll.Head && nodeToInsert == dll.Tail { //nolint:gocritic
		return
	}

	dll.Remove(nodeToInsert)
	nodeToInsert.Prev = node.Prev
	nodeToInsert.Next = node
	if node.Prev.Next == nil {
		dll.Head = nodeToInsert
	} else {
		node.Prev.Next = nodeToInsert
	}
	node.Prev = nodeToInsert
}

// InsertAfter inserts node after the provided node
func (dll *DoublyLinkedList) InsertAfter(node, nodeToInsert *DllNode) {
	if nodeToInsert == dll.Head && nodeToInsert == dll.Tail { //nolint:gocritic
		return
	}

	dll.Remove(nodeToInsert)
	nodeToInsert.Prev = node
	nodeToInsert.Next = node.Next
	if node.Next == nil {
		dll.Tail = nodeToInsert
	} else {
		node.Next.Prev = nodeToInsert
	}
	node.Next = nodeToInsert
}

// Remove removes the node
func (dll *DoublyLinkedList) Remove(node *DllNode) {
	if node == dll.Head {
		dll.Head = dll.Head.Next
	}
	if node == dll.Tail {
		dll.Tail = dll.Tail.Prev
	}
	dll.removeNode(node)
}

// Reset resets the whole DLL
func (dll *DoublyLinkedList) Reset() {
	cNode := dll.Head
	for cNode != nil {
		cNode.Vertex.NumOfDeps = len(cNode.Vertex.Dependencies)
		cNode.Vertex.Visiting = false
		cNode.Vertex.Visited = false
		cNode = cNode.Next
	}
}

func (dll *DoublyLinkedList) removeNode(node *DllNode) {
	if node.Prev != nil {
		node.Prev.Next = node.Next
	}
	if node.Next != nil {
		node.Next.Prev = node.Prev
	}
	node.Prev = nil
	node.Next = nil
}
