package lists

import (
	"sync"
)

type Queue[T any] interface {
	Read() (T, bool)
	WriteAll([]T)
	Delete()
}

// The DFSQueue behaves somewhat like a stack. The only
// difference is that if the stack is full, newer elements
// are written to the top of the stack.
type DFSQueue[T any] struct {
	arr       []T
	front     int
	back      int
	isFull    bool
	deleted   bool
	lock      sync.Mutex
	writeCond sync.Cond
}

func NewDFSQueue[T any](size int) *DFSQueue[T] {
	cq := &DFSQueue[T]{
		arr:     make([]T, size),
		back:    0, //Push and pop from the back.
		front:   0,
		isFull:  false,
		deleted: false,
	}
	cq.writeCond.L = &cq.lock
	return cq
}

func (cq *DFSQueue[T]) Read() (val T, deleted bool) {
	cq.lock.Lock()
	defer cq.lock.Unlock()

	if cq.deleted {
		return val, true
	}
	if cq.back == cq.front && !cq.isFull {
		cq.writeCond.Wait()
	}
	if cq.deleted {
		return val, true
	}
	cq.back = ((cq.back - 1) + len(cq.arr)) % len(cq.arr)
	val = cq.arr[cq.back]
	return val, false
}

func (cq *DFSQueue[T]) Delete() {
	cq.lock.Lock()
	defer cq.lock.Unlock()

	cq.deleted = true
	cq.writeCond.Broadcast()
}

// WriteAll will overwrite the elements that are at the end of the
// queue.
func (cq *DFSQueue[T]) WriteAll(newElements []T) {
	cq.lock.Lock()
	defer cq.lock.Unlock()

	for i := len(newElements) - 1; i >= 0; i-- {
		cq.arr[cq.back] = newElements[i]
		cq.back = (cq.back + 1) % len(cq.arr)
		if cq.isFull {
			cq.front = (cq.front + 1) % len(cq.arr)
		}
		if cq.back == cq.front {
			cq.isFull = true
		}
		cq.writeCond.Signal()
	}
}

// The BFSQueue behaves like a regular queue. The only
// difference is that if the queue is full, newer elements
// are DISCARDED.
type BFSQueue[T any] struct {
	arr       []T
	front     int
	back      int
	isFull    bool
	deleted   bool
	lock      sync.Mutex
	writeCond sync.Cond
}

func NewBFSQueue[T any](size int) *BFSQueue[T] {
	cq := &BFSQueue[T]{
		arr:     make([]T, size),
		front:   0,
		back:    0,
		isFull:  false,
		deleted: false,
	}
	cq.writeCond.L = &cq.lock
	return cq
}

func (cq *BFSQueue[T]) Read() (val T, deleted bool) {
	cq.lock.Lock()
	defer cq.lock.Unlock()

	if cq.deleted {
		return val, true
	}
	if cq.front == cq.back && !cq.isFull {
		cq.writeCond.Wait()
	}
	if cq.deleted {
		return val, true
	}
	val = cq.arr[cq.front]
	cq.front = (cq.front + 1) % len(cq.arr)
	if cq.isFull {
		cq.isFull = false
	}
	return val, false
}

func (cq *BFSQueue[T]) WriteAll(vals []T) {
	cq.lock.Lock()
	defer cq.lock.Unlock()

	idx := 0
	for !cq.isFull && idx < len(vals) {
		cq.write(vals[idx])
		idx++
	}
}

func (cq *BFSQueue[T]) Delete() {
	cq.lock.Lock()
	defer cq.lock.Unlock()

	cq.deleted = true
	cq.writeCond.Broadcast()
}

func (cq *BFSQueue[T]) write(val T) {
	cq.arr[cq.back] = val
	cq.back = (cq.back + 1) % len(cq.arr)
	if cq.back == cq.front {
		cq.isFull = true
	}
	cq.writeCond.Signal()
}
