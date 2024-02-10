package caches

import "fmt"

var EmptyList error = fmt.Errorf("List is empty\n")

// linkedList operations are not thread safe.
type linkedList[K any, T any] struct {
	sentinel *listNode[K, T]
	size     int
}

type listNode[K any, T any] struct {
	key        K
	value      T
	next, prev *listNode[K, T]
}

func newLinkedList[K any, T any]() *linkedList[K, T] {
	sentinel := &listNode[K, T]{}
	sentinel.next = sentinel
	sentinel.prev = sentinel
	return &linkedList[K, T]{
		sentinel: sentinel,
		size:     0,
	}
}

func (ll *linkedList[K, T]) addToFront(key K, val T) *listNode[K, T] {
	ll.size++
	node := &listNode[K, T]{
		key:   key,
		value: val,
		next:  ll.sentinel.next,
		prev:  ll.sentinel,
	}

	ll.sentinel.next.prev = node
	ll.sentinel.next = node

	return node
}

func (ll *linkedList[K, T]) addToBack(key K, val T) *listNode[K, T] {
	ll.size++
	node := &listNode[K, T]{
		key:   key,
		value: val,
		next:  ll.sentinel,
		prev:  ll.sentinel.prev,
	}
	ll.sentinel.prev.next = node
	ll.sentinel.prev = node
	return node
}

func (ll *linkedList[K, T]) moveToFront(node *listNode[K, T]) {
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

func (ll *linkedList[K, T]) remove(node *listNode[K, T]) {
	ll.size--
	node.prev.next = node.next
	node.next.prev = node.prev
	node.next = nil
	node.prev = nil
}

func (ll *linkedList[K, T]) popBack() (*listNode[K, T], error) {
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

func (ll *linkedList[K, T]) popFront() (*listNode[K, T], error) {
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

func (ll *linkedList[K, T]) Len() int {
	return ll.size
}
