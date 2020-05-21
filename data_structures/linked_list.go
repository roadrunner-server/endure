package data_structures

// Node of the DoublyLL
type Node struct {
	Value      int
	Prev, Next *Node
}

// DoublyLinkedList is the node of DLL which is connected to the tail and the head
type DoublyLinkedList struct {
	Head, Tail *Node
}

func NewDoublyLinkedList() *DoublyLinkedList {
	return nil
}

func (dll *DoublyLinkedList) SetHead(node *Node) {

}

func (dll *DoublyLinkedList) SetTail(node *Node) {
}

func (dll *DoublyLinkedList) InsertBefore(node, nodeToInsert *Node) {
}

func (dll *DoublyLinkedList) InsertAfter(node, nodeToInsert *Node) {
}

func (dll *DoublyLinkedList) InsertAtPosition(position int, nodeToInsert *Node) {
}

func (dll *DoublyLinkedList) RemoveNodesWithValue(value int) {
}

func (dll *DoublyLinkedList) Remove(node *Node) {
}

func (dll *DoublyLinkedList) ContainsNodeWithValue(value int) bool {
	return false
}

