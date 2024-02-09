package caches_test

import (
	"github.com/adityachandla/graph_access_service/caches"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestLruCase(t *testing.T) {
	lrfu := caches.NewLrfuCache[int, int](2, 1.0)
	lrfu.Put(1, 101)
	lrfu.Put(2, 102)
	lrfu.Get(1)
	lrfu.Get(1)
	lrfu.Get(2)
	//1 should be evicted
	lrfu.Put(3, 103)
	_, found := lrfu.Get(1)
	assert.False(t, found)
	_, found = lrfu.Get(2)
	assert.True(t, found)
}

func TestLfuCase(t *testing.T) {
	lrfu := caches.NewLrfuCache[int, int](2, 0.0)
	lrfu.Put(1, 101)
	lrfu.Put(2, 102)
	lrfu.Get(1)
	lrfu.Get(1)
	lrfu.Get(2)
	//2 should be evicted
	lrfu.Put(3, 103)
	_, found := lrfu.Get(1)
	assert.True(t, found)
	_, found = lrfu.Get(2)
	assert.False(t, found)
}

func TestMiddleCase(t *testing.T) {
	lrfu := caches.NewLrfuCache[int, int](2, 0.4)
	lrfu.Put(1, 101)
	lrfu.Put(2, 102)
	lrfu.Get(1)
	lrfu.Get(1)
	lrfu.Get(2)
	//2 should be evicted
	lrfu.Put(3, 103)
	_, found := lrfu.Get(1)
	assert.True(t, found)
	_, found = lrfu.Get(2)
	assert.False(t, found)
}

func TestMiddleCase2(t *testing.T) {
	lrfu := caches.NewLrfuCache[int, int](2, 0.6)
	lrfu.Put(1, 101)
	lrfu.Put(2, 102)
	lrfu.Get(1)
	lrfu.Get(1)
	lrfu.Get(2)
	//2 should be evicted
	lrfu.Put(3, 103)
	_, found := lrfu.Get(1)
	assert.False(t, found)
	_, found = lrfu.Get(2)
	assert.True(t, found)
}

func BenchmarkLrfu_Put(b *testing.B) {
	lrfu := caches.NewLrfuCache[int, int](1000, 0.5)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		lrfu.Put(i%1000, i+1)
	}
}

func BenchmarkLrfu_Get(b *testing.B) {
	lrfu := caches.NewLrfuCache[int, int](1000, 0.5)
	for i := 0; i < 1000; i++ {
		lrfu.Put(i, 1000+i)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		lrfu.Get(i % 1000)
	}
}
