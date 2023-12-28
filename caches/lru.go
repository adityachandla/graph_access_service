package caches

import "fmt"

// This interface specifies how the cache is
// supposed to fetch a key if it is not present
// in the cache.
type Fetcher[K comparable, T any] interface {
	Fetch(name K) T
}

type LRU[K comparable, V any] struct {
	mapping      map[K]*listNode[V]
	revMapping   map[*listNode[V]]K
	recencyQueue *LinkedList[V]
	fetcher      Fetcher[K, V] //This interface will fetch the
	maxSize      int
}

func NewLRU[K comparable, V any](fetcher Fetcher[K, V], maxSize int) *LRU[K, V] {
	if maxSize < 1 {
		panic("LRU maxSize should be >= 1")
	}
	return &LRU[K, V]{
		mapping:      make(map[K]*listNode[V]),
		revMapping:   make(map[*listNode[V]]K),
		recencyQueue: NewLinkedList[V](),
		fetcher:      fetcher,
		maxSize:      maxSize,
	}
}

func (lru *LRU[K, V]) Get(key K) V {
	if ref, ok := lru.mapping[key]; ok {
		//Present in the map
		lru.recencyQueue.MoveToFront(ref)
		return ref.value
	}
	//Not present in the map
	val := lru.fetcher.Fetch(key)
	valRef := lru.recencyQueue.AddToFront(val)
	lru.mapping[key] = valRef
	//Size limit exceeded
	if lru.recencyQueue.Size() > lru.maxSize {
		toDeleteRef, err := lru.recencyQueue.PopBack()
		if err != nil {
			panic(err)
		}
		toDeleteKey := lru.revMapping[toDeleteRef]
		delete(lru.mapping, toDeleteKey)
		delete(lru.revMapping, toDeleteRef)
	}
	return val
}

var EmptyList error = fmt.Errorf("List is empty")

type LinkedList[T any] struct {
	start, end *listNode[T]
	size       int
}

type listNode[T any] struct {
	value      T
	next, prev *listNode[T]
}

func NewLinkedList[T any]() *LinkedList[T] {
	return &LinkedList[T]{
		start: nil,
		end:   nil,
		size:  0,
	}
}

func (ll *LinkedList[T]) AddToFront(val T) *listNode[T] {
	node := &listNode[T]{value: val, next: nil, prev: nil}
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

func (ll *LinkedList[T]) MoveToFront(node *listNode[T]) {
	if node.prev == nil {
		//Already the first node
		return
	}
	// Manage prev-next pointers
	node.prev.next = node.next
	if node.next != nil {
		node.next.prev = node.prev
	}
	//Move to front
	node.prev = nil
	node.next = ll.start
	ll.start.prev = node
	ll.start = node
}

func (ll *LinkedList[T]) PopBack() (*listNode[T], error) {
	if ll.end == nil {
		return nil, EmptyList
	}
	ll.size--
	node := ll.end
	ll.end = ll.end.prev
	return node, nil
}

func (ll *LinkedList[T]) Size() int {
	return ll.size
}
