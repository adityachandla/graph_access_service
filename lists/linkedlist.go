package lists

import "fmt"

var EmptyList error = fmt.Errorf("List is empty\n")

// LinkedList operations are not thread safe.
type LinkedList[K any, T any] struct {
	sentinel *ListNode[K, T]
	size     int
}

type ListNode[K any, T any] struct {
	Key        K
	Value      T
	next, prev *ListNode[K, T]
}

func NewLinkedList[K any, T any]() *LinkedList[K, T] {
	sentinel := &ListNode[K, T]{}
	sentinel.next = sentinel
	sentinel.prev = sentinel
	return &LinkedList[K, T]{
		sentinel: sentinel,
		size:     0,
	}
}

func (ll *LinkedList[K, T]) AddToFront(key K, val T) *ListNode[K, T] {
	ll.size++
	node := &ListNode[K, T]{
		Key:   key,
		Value: val,
		next:  ll.sentinel.next,
		prev:  ll.sentinel,
	}

	ll.sentinel.next.prev = node
	ll.sentinel.next = node

	return node
}

func (ll *LinkedList[K, T]) AddToBack(key K, val T) *ListNode[K, T] {
	ll.size++
	node := &ListNode[K, T]{
		Key:   key,
		Value: val,
		next:  ll.sentinel,
		prev:  ll.sentinel.prev,
	}
	ll.sentinel.prev.next = node
	ll.sentinel.prev = node
	return node
}

func (ll *LinkedList[K, T]) MoveToFront(node *ListNode[K, T]) {
	if node.prev == ll.sentinel {
		//Already the first node
		return
	}
	// Manage prev-next nodes' pointers
	node.prev.next = node.next
	node.next.prev = node.prev

	//Move to front
	node.prev = ll.sentinel
	node.next = ll.sentinel.next

	ll.sentinel.next.prev = node
	ll.sentinel.next = node
}

func (ll *LinkedList[K, T]) Remove(node *ListNode[K, T]) {
	ll.size--
	node.prev.next = node.next
	node.next.prev = node.prev
	node.next = nil
	node.prev = nil
}

func (ll *LinkedList[K, T]) PopBack() (*ListNode[K, T], error) {
	if ll.size == 0 {
		return nil, EmptyList
	}
	ll.size--
	node := ll.sentinel.prev
	ll.sentinel.prev = node.prev
	node.prev.next = ll.sentinel

	//Avoid memory leak
	node.prev = nil
	node.next = nil
	return node, nil
}

func (ll *LinkedList[K, T]) PopFront() (*ListNode[K, T], error) {
	if ll.size == 0 {
		return nil, EmptyList
	}
	ll.size--
	node := ll.sentinel.next

	ll.sentinel.next = node.next
	node.next.prev = ll.sentinel

	//Avoid memory leak
	node.prev = nil
	node.next = nil
	return node, nil
}

func (ll *LinkedList[K, T]) Len() int {
	return ll.size
}
