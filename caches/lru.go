package caches

import (
	"fmt"
	"sync"
)

// This interface specifies how the cache is
// supposed to fetch a key if it is not present
// in the cache.
type Fetcher[K comparable, T any] interface {
	Fetch(name K) T
}

type LRU[K comparable, V any] struct {
	mapping      map[K]*listNode[K, V]
	recencyQueue *LinkedList[K, V]
	fetcher      Fetcher[K, V]
	maxSize      int
	lock         *sync.Mutex
}

func NewLRU[K comparable, V any](fetcher Fetcher[K, V], maxSize int) *LRU[K, V] {
	if maxSize < 1 {
		panic("LRU maxSize should be >= 1")
	}
	return &LRU[K, V]{
		mapping:      make(map[K]*listNode[K, V]),
		recencyQueue: NewLinkedList[K, V](),
		fetcher:      fetcher,
		maxSize:      maxSize,
		lock:         &sync.Mutex{},
	}
}

// Get
// This function does not lock for the entire duration
// because the fetch operation might be expensive.
// As a result, it is possible that two goroutines might
// call fetch on the same key.
func (lru *LRU[K, V]) Get(key K) V {
	lru.lock.Lock()
	defer lru.lock.Unlock()
	if cached, ok := lru.getFromCache(key); ok {
		return cached
	}
	// Not present in the map
	// Do the fetch operation after releasing the lock.
	lru.lock.Unlock()
	val := lru.fetcher.Fetch(key)
	lru.lock.Lock()
	// Some other goroutine might have fetched this key in
	// the meantime.
	if cached, ok := lru.getFromCache(key); ok {
		return cached
	}
	lru.addToCache(key, val)
	if lru.recencyQueue.Size() > lru.maxSize {
		lru.evictLast()
	}
	return val
}

func (lru *LRU[K, V]) addToCache(key K, val V) {
	valRef := lru.recencyQueue.AddToFront(key, val)
	lru.mapping[key] = valRef
}

func (lru *LRU[K, V]) getFromCache(key K) (val V, ok bool) {
	if ref, ok := lru.mapping[key]; ok {
		lru.recencyQueue.MoveToFront(ref)
		return ref.value, true
	}
	return
}

func (lru *LRU[K, V]) evictLast() {
	toDeleteRef, err := lru.recencyQueue.PopBack()
	if err != nil {
		panic(err)
	}
	delete(lru.mapping, toDeleteRef.key)
}

var EmptyList error = fmt.Errorf("List is empty")

// Operations on the linked list are not thread safe.
type LinkedList[K any, T any] struct {
	start, end *listNode[K, T]
	size       int
}

type listNode[K any, T any] struct {
	key        K
	value      T
	next, prev *listNode[K, T]
}

func NewLinkedList[K any, T any]() *LinkedList[K, T] {
	return &LinkedList[K, T]{
		start: nil,
		end:   nil,
		size:  0,
	}
}

func (ll *LinkedList[K, T]) AddToFront(key K, val T) *listNode[K, T] {
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

func (ll *LinkedList[K, T]) MoveToFront(node *listNode[K, T]) {
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

func (ll *LinkedList[K, T]) PopBack() (*listNode[K, T], error) {
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

func (ll *LinkedList[K, T]) Size() int {
	return ll.size
}
