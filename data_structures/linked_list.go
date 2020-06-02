package data_structures

// Vertex of the DoublyLL
type DllNode struct {
	Value      int
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
func (dll *DoublyLinkedList) InsertAtPosition(position int, nodeToInsert *DllNode) {
	if position == 1 {
		dll.SetHead(nodeToInsert)
		return
	}
	node := dll.Head
	currentPos := 1
	for node != nil && currentPos != position {
		node = node.Next
		currentPos += 1
	}
	if node != nil {
		dll.InsertBefore(node, nodeToInsert)
	} else {
		dll.SetTail(nodeToInsert)
	}
}

func (dll *DoublyLinkedList) RemoveNodesWithValue(value int) {
	node := dll.Head
	for node != nil {
		nodeToRemove := node
		node = node.Next
		if nodeToRemove.Value == value {
			dll.Remove(node)
		}
	}
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

func (dll *DoublyLinkedList) ContainsNodeWithValue(value int) bool {
	node := dll.Head
	for node != nil && node.Value != value {
		node = node.Next
	}
	return node != nil
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
