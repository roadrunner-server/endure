package stack

import (
	"sync/atomic"
	"unsafe"
)

type Stack[T any] struct {
	head *Node[T]
	size uint64
}

type Node[T any] struct {
	value T
	next  *Node[T]
}

func New[T any]() *Stack[T] {
	return &Stack[T]{}
}

func (s *Stack[T]) Len() uint64 {
	return atomic.LoadUint64(&s.size)
}

func (s *Stack[T]) Push(value T) {
	newNode := &Node[T]{value: value}
	for {
		head := s.head
		newNode.next = head
		if atomic.CompareAndSwapPointer((*unsafe.Pointer)(unsafe.Pointer(&s.head)), unsafe.Pointer(head), unsafe.Pointer(newNode)) {
			atomic.AddUint64(&s.size, 1)
			return
		}
	}
}

func (s *Stack[T]) Pop() (T, bool) {
	for {
		head := s.head
		if head == nil {
			var placeholder T
			return placeholder, false
		}
		if atomic.CompareAndSwapPointer((*unsafe.Pointer)(unsafe.Pointer(&s.head)), unsafe.Pointer(head), unsafe.Pointer(head.next)) {
			atomic.AddUint64(&s.size, ^uint64(0))
			return head.value, true
		}
	}
}
