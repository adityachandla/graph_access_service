package caches

import "fmt"

var EmptyList error = fmt.Errorf("List is empty\n")

// linkedList operations are not thread safe.
type linkedList[K any, T any] struct {
	start, end *listNode[K, T]
	size       int
}

type listNode[K any, T any] struct {
	key        K
	value      T
	next, prev *listNode[K, T]
}

func newLinkedList[K any, T any]() *linkedList[K, T] {
	return &linkedList[K, T]{
		start: nil,
		end:   nil,
		size:  0,
	}
}

func (ll *linkedList[K, T]) addToFront(key K, val T) *listNode[K, T] {
	node := &listNode[K, T]{key: key, value: val, next: nil, prev: nil}
	ll.size++
	if ll.start != nil {
		node.next = ll.start
		ll.start.prev = node
	} else {
		//First insertion
		ll.end = node
	}
	ll.start = node
	return node
}

func (ll *linkedList[K, T]) moveToFront(node *listNode[K, T]) {
	if node.prev == nil {
		//Already the first node
		return
	}
	// Manage prev-next nodes' pointers
	node.prev.next = node.next
	if node.next != nil {
		node.next.prev = node.prev
	}
	if ll.end == node {
		ll.end = node.prev
	}
	//Move to front
	node.prev = nil
	node.next = ll.start
	ll.start.prev = node
	ll.start = node
}

func (ll *linkedList[K, T]) popBack() (*listNode[K, T], error) {
	if ll.end == nil {
		return nil, EmptyList
	}
	ll.size--
	node := ll.end
	ll.end = ll.end.prev
	//These are important to avoid memory leak
	node.prev = nil
	node.next = nil
	return node, nil
}

func (ll *linkedList[K, T]) Len() int {
	return ll.size
}
