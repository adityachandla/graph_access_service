package caches_test

import (
	"testing"

	"github.com/adityachandla/graph_access_service/caches"
)

type dummyFetcher struct {
	invokeCount int
}

func (d *dummyFetcher) Fetch(in int) int {
	d.invokeCount++
	return in + 10
}

func TestFetchFromCache(t *testing.T) {
	fetcher := &dummyFetcher{}
	lru := caches.NewLRU[int, int](fetcher, 4)
	lru.Get(22)
	lru.Get(22)
	checkInvoke(t, fetcher, 1)
}

func TestEviction(t *testing.T) {
	fetcher := &dummyFetcher{}
	lru := caches.NewLRU[int, int](fetcher, 2)
	lru.Get(22)
	lru.Get(23)
	lru.Get(24)
	lru.Get(22)
	checkInvoke(t, fetcher, 4)
}

func TestReuseLast(t *testing.T) {
	fetcher := &dummyFetcher{}
	lru := caches.NewLRU[int, int](fetcher, 3)
	lru.Get(22) //invoke
	lru.Get(23) //invoke
	lru.Get(24) //invoke
	lru.Get(22) //Now 23 should be the last one.
	lru.Get(88) //invoke, 23 would be evicted.
	lru.Get(23) //invoke.
	checkInvoke(t, fetcher, 5)
}

func TestReuseMiddle(t *testing.T) {
	fetcher := &dummyFetcher{}
	lru := caches.NewLRU[int, int](fetcher, 3)
	lru.Get(22) //invoke
	lru.Get(23) //invoke
	lru.Get(24) //invoke
	lru.Get(23) //Now 22 should be the last one.
	lru.Get(88) //invoke, 22 would be evicted.
	lru.Get(99) //invoke, 24 would be evicted.
	lru.Get(23) //Should be fetched from cache
	checkInvoke(t, fetcher, 5)
}

func checkInvoke(t *testing.T, fetcher *dummyFetcher, count int) {
	if fetcher.invokeCount != count {
		t.Fatalf("Expected %d invocations got %d\n", count, fetcher.invokeCount)
	}
}
