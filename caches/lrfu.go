package caches

import (
	"math"
	"sync"
)

const BASE = 0.5

type Lrfu[K comparable, V any] struct {
	mapping map[K]*heapNode[K, V]
	// This heap contains the min value at the top.
	heap         []*heapNode[K, V]
	lock         sync.Mutex
	maxSize      int
	combinedBase float64
	time         uint32
}

func NewLrfuCache[K comparable, V any](size int, lambda float64) *Lrfu[K, V] {
	return &Lrfu[K, V]{
		mapping:      make(map[K]*heapNode[K, V]),
		heap:         make([]*heapNode[K, V], 0),
		lock:         sync.Mutex{},
		maxSize:      size,
		combinedBase: math.Pow(BASE, lambda),
		time:         0,
	}
}

func (lrfu *Lrfu[K, V]) Put(key K, value V) {
	lrfu.lock.Lock()
	defer lrfu.lock.Unlock()
	lrfu.time++

	newNode := heapNode[K, V]{
		key:           key,
		value:         value,
		timeAccessed:  lrfu.time,
		combinedScore: lrfu.combinedScore(0),
	}
	lrfu.mapping[key] = &newNode
	if lrfu.maxSize == len(lrfu.heap) {
		keyToEvict := lrfu.heap[0].key
		delete(lrfu.mapping, keyToEvict)
		newNode.nodeIdx = 0
		lrfu.heap[0] = &newNode
		lrfu.floatDown(0)
	} else {
		lrfu.heap = append(lrfu.heap, &newNode)
		newNode.nodeIdx = len(lrfu.heap) - 1
		lrfu.floatUp(len(lrfu.heap) - 1)
	}
}

func (lrfu *Lrfu[K, V]) Present(key K) bool {
	lrfu.lock.Lock()
	defer lrfu.lock.Unlock()
	_, ok := lrfu.mapping[key]
	return ok
}

func (lrfu *Lrfu[K, V]) Get(key K) (V, bool) {
	lrfu.lock.Lock()
	defer lrfu.lock.Unlock()
	lrfu.time++

	//After access, the value may have increased, so we need to
	//float it down.
	if nodePtr, ok := lrfu.mapping[key]; ok {
		nodePtr.combinedScore = lrfu.combinedScore(0) +
			(lrfu.combinedScore(lrfu.time-nodePtr.timeAccessed) * nodePtr.combinedScore)
		nodePtr.timeAccessed = lrfu.time
		lrfu.floatDown(nodePtr.nodeIdx)
		return nodePtr.value, true
	}
	var v V
	return v, false
}

func (lrfu *Lrfu[K, V]) floatUp(index int) {
	for index > 0 {
		parent := (index - 1) / 2
		if lrfu.compare(parent, index) > 0 {
			lrfu.swapInHeap(index, parent)
		} else {
			break
		}
		index = parent
	}
}

func (lrfu *Lrfu[K, V]) floatDown(index int) {
	for (2*index)+1 < len(lrfu.heap) {
		leftChildIdx := (2 * index) + 1
		rightChildIdx := (2 * index) + 2
		minIdx := index
		if leftChildIdx < len(lrfu.heap) && lrfu.compare(leftChildIdx, minIdx) < 0 {
			minIdx = leftChildIdx
		}
		if rightChildIdx < len(lrfu.heap) && lrfu.compare(rightChildIdx, minIdx) < 0 {
			minIdx = rightChildIdx
		}
		if minIdx == index {
			return
		}
		lrfu.swapInHeap(index, minIdx)
		index = minIdx
	}
}

func (lrfu *Lrfu[K, V]) swapInHeap(one, two int) {
	lrfu.heap[one].nodeIdx, lrfu.heap[two].nodeIdx = two, one
	lrfu.heap[one], lrfu.heap[two] = lrfu.heap[two], lrfu.heap[one]
}

func (lrfu *Lrfu[K, V]) compare(one, two int) int {
	oneCrf := lrfu.heap[one].combinedScore * lrfu.combinedScore(lrfu.time-lrfu.heap[one].timeAccessed)
	twoCrf := lrfu.heap[two].combinedScore * lrfu.combinedScore(lrfu.time-lrfu.heap[two].timeAccessed)
	if oneCrf > twoCrf {
		return 1
	} else if oneCrf < twoCrf {
		return -1
	}
	return 0
}

func (lrfu *Lrfu[K, V]) combinedScore(x uint32) float64 {
	return math.Pow(lrfu.combinedBase, float64(x))
}

type heapNode[K comparable, V any] struct {
	key           K
	value         V
	nodeIdx       int
	timeAccessed  uint32
	combinedScore float64
}
