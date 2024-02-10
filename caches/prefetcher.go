package caches

import "sync"

type PrefetchCache[K comparable, V any] struct {
	elementMap  map[K]*listNode[K, V]
	list        *linkedList[K, V]
	numElements int
	maxSize     int
	lock        sync.Mutex
}

func NewPrefetchCache[K comparable, V any](size int) *PrefetchCache[K, V] {
	if size <= 0 {
		panic("Size of cache needs to be greater than 0.")
	}
	return &PrefetchCache[K, V]{
		elementMap:  make(map[K]*listNode[K, V]),
		list:        newLinkedList[K, V](),
		numElements: 0,
		maxSize:     size,
	}
}

func (pc *PrefetchCache[K, V]) Get(key K) (val V, found bool) {
	pc.lock.Lock()
	defer pc.lock.Unlock()
	if node, ok := pc.elementMap[key]; ok {
		val = node.value
		delete(pc.elementMap, key)
		pc.list.remove(node)
		pc.numElements--
		return val, true
	}
	return val, false
}

func (pc *PrefetchCache[K, V]) Put(key K, value V) {
	pc.lock.Lock()
	defer pc.lock.Unlock()
	node := pc.list.addToBack(key, value)
	pc.elementMap[key] = node
	if pc.numElements == pc.maxSize {
		toRemove, _ := pc.list.popFront()
		delete(pc.elementMap, toRemove.key)
	} else {
		pc.numElements++
	}
}

func (pc *PrefetchCache[K, V]) Len() int {
	return pc.numElements
}
