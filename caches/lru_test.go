package caches_test

import (
	"github.com/stretchr/testify/assert"
	"testing"

	"github.com/adityachandla/graph_access_service/caches"
)

func TestFetchFromCache(t *testing.T) {
	lru := caches.NewLRU[int, int](4)
	lru.Put(22, 23)
	_, found := lru.Get(22)
	assert.True(t, found)
}

func TestEviction(t *testing.T) {
	lru := caches.NewLRU[int, int](2)
	lru.Put(22, 101)
	lru.Put(23, 102)
	lru.Put(24, 103)
	_, found := lru.Get(22)
	assert.False(t, found)
	_, found = lru.Get(24)
	assert.True(t, found)
	_, found = lru.Get(23)
	assert.True(t, found)
}

func TestReuseLast(t *testing.T) {
	lru := caches.NewLRU[int, int](3)
	lru.Put(22, 101)        //invoke
	lru.Put(23, 102)        //invoke
	lru.Put(24, 103)        //invoke
	lru.Get(22)             //Now 23 should be the last one.
	lru.Put(88, 104)        //invoke, 23 would be evicted.
	_, found := lru.Get(23) //invoke.
	assert.False(t, found)
}

func TestReuseMiddle(t *testing.T) {
	lru := caches.NewLRU[int, int](3)
	lru.Put(22, 101) //invoke
	lru.Put(23, 102) //invoke
	lru.Put(24, 103) //invoke
	lru.Get(23)      //Now 22 should be the last one.
	lru.Put(88, 104) //invoke, 22 would be evicted.
	_, found := lru.Get(22)
	assert.False(t, found)
	lru.Put(99, 105) //invoke, 24 would be evicted.
	_, found = lru.Get(24)
	assert.False(t, found)
	_, found = lru.Get(23) //Should be fetched from cache
	assert.True(t, found)
}
