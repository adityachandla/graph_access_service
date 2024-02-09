package caches

import (
	"sync"
)

type LRU[K comparable, V any] struct {
	mapping      map[K]*listNode[K, V]
	recencyQueue *linkedList[K, V]
	maxSize      int
	lock         *sync.Mutex
}

func NewLRU[K comparable, V any](maxSize int) *LRU[K, V] {
	if maxSize < 1 {
		panic("LRU maxSize should be >= 1")
	}
	return &LRU[K, V]{
		mapping:      make(map[K]*listNode[K, V]),
		recencyQueue: newLinkedList[K, V](),
		maxSize:      maxSize,
		lock:         &sync.Mutex{},
	}
}

func (lru *LRU[K, V]) Get(key K) (val V, ok bool) {
	lru.lock.Lock()
	defer lru.lock.Unlock()
	if cached, ok := lru.getFromCache(key); ok {
		return cached, true
	}
	return
}

func (lru *LRU[K, V]) Put(key K, value V) {
	lru.addToCache(key, value)
	if lru.recencyQueue.Len() > lru.maxSize {
		lru.evictLast()
	}
}

func (lru *LRU[K, V]) addToCache(key K, val V) {
	valRef := lru.recencyQueue.addToFront(key, val)
	lru.mapping[key] = valRef
}

func (lru *LRU[K, V]) getFromCache(key K) (val V, ok bool) {
	if ref, ok := lru.mapping[key]; ok {
		lru.recencyQueue.moveToFront(ref)
		return ref.value, true
	}
	return
}

func (lru *LRU[K, V]) evictLast() {
	toDeleteRef, err := lru.recencyQueue.popBack()
	if err != nil {
		panic(err)
	}
	delete(lru.mapping, toDeleteRef.key)
}
