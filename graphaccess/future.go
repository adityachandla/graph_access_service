package graphaccess

import "sync"

type future[T any] struct {
	val T
	wg  sync.WaitGroup
}

func newFuture[T any]() *future[T] {
	f := &future[T]{}
	f.wg.Add(1)
	return f
}

func (f *future[T]) put(val T) {
	f.val = val
	f.wg.Done()
}

func (f *future[T]) get() T {
	f.wg.Wait()
	return f.val
}
