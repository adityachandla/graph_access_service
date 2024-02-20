package lists

import (
	"sync"
)

type CircularQueue[T any] struct {
	arr       []T
	front     int
	back      int
	isFull    bool
	lock      sync.Mutex
	writeCond sync.Cond
}

func NewCircularQueue[T any](size int) *CircularQueue[T] {
	cq := &CircularQueue[T]{
		arr:    make([]T, size),
		front:  0,
		back:   0,
		isFull: false,
	}
	cq.writeCond.L = &cq.lock
	return cq
}

func (cq *CircularQueue[T]) Read() T {
	cq.lock.Lock()
	defer cq.lock.Unlock()

	if cq.front == cq.back && !cq.isFull {
		cq.writeCond.Wait()
	}
	val := cq.arr[cq.front]
	cq.front = (cq.front + 1) % len(cq.arr)
	if cq.isFull {
		cq.isFull = false
	}
	return val
}

func (cq *CircularQueue[T]) Write(newElements []T) {
	cq.lock.Lock()
	defer cq.lock.Unlock()

	for i := len(newElements) - 1; i >= 0; i-- {
		cq.writeOrOverwrite(newElements[i])
	}
}

func (cq *CircularQueue[T]) writeOrOverwrite(element T) {
	cq.arr[cq.back] = element
	cq.back = (cq.back + 1) % len(cq.arr)
	if cq.isFull {
		//Overwrite case
		cq.front = (cq.front + 1) % len(cq.arr)
		return
	}
	if cq.back == cq.front {
		cq.isFull = true
	}
	cq.writeCond.Signal()
}
