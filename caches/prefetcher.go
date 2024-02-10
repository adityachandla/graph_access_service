package caches

type PrefetchCache[K comparable, V any] struct {
	elementMap  map[K]*listNode[K, V]
	list        *linkedList[K, V]
	numElements int
	maxSize     int
}

func NewPrefetchCache[K comparable, V any](size int) *PrefetchCache[K, V] {
	return &PrefetchCache[K, V]{
		elementMap:  make(map[K]*listNode[K, V]),
		list:        newLinkedList[K, V](),
		numElements: 0,
		maxSize:     size,
	}
}

// TODO implement these
func (pc *PrefetchCache[K, V]) Get(key K) (val V, ok bool) {
	return
}

func (pc *PrefetchCache[K, V]) Put(key K, value V) {
}
