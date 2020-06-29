package structures

// Vertex of the DoublyLL
type DllNode struct {
	Vertex     *Vertex
	Prev, Next *DllNode
}

// DoublyLinkedList is the node of DLL which is connected to the tail and the head
type DoublyLinkedList struct {
	Head, Tail *DllNode
}

func NewDoublyLinkedList() *DoublyLinkedList {
	return &DoublyLinkedList{}
}

// O(1) time + space
func (dll *DoublyLinkedList) SetHead(node *DllNode) {
	if dll.Head == nil {
		dll.Head = node
		dll.Tail = node
		return
	}
	dll.InsertBefore(dll.Head, node)
}

func (dll *DoublyLinkedList) Push(vertex *Vertex) {
	node := &DllNode{
		Vertex: vertex,
	}
	node.Next = dll.Head
	dll.Head.Prev = node

	node.Prev = nil

	dll.Head = node
}

func (dll *DoublyLinkedList) PushTail(vertex *Vertex) {
	node := &DllNode{
		Vertex: vertex,
	}
	node.Next = dll.Tail
	dll.Tail.Next = node

	node.Prev = dll.Head

	dll.Tail = node
}

// constant O(1) time and space
func (dll *DoublyLinkedList) SetTail(node *DllNode) {
	if dll.Tail == nil {
		dll.SetHead(node)
	}
	dll.InsertAfter(dll.Tail, node)
}

func (dll *DoublyLinkedList) InsertBefore(node, nodeToInsert *DllNode) {
	if nodeToInsert == dll.Head && nodeToInsert == dll.Tail {
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

func (dll *DoublyLinkedList) InsertAfter(node, nodeToInsert *DllNode) {
	if nodeToInsert == dll.Head && nodeToInsert == dll.Tail {
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


func (dll *DoublyLinkedList) Remove(node *DllNode) {
	if node == dll.Head {
		dll.Head = dll.Head.Next
	}
	if node == dll.Tail {
		dll.Tail = dll.Tail.Prev
	}
	dll.removeNode(node)
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
