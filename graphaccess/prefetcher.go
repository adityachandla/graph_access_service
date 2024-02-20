package graphaccess

import (
	"github.com/adityachandla/graph_access_service/caches"
	"github.com/adityachandla/graph_access_service/lists"
	"sync"
)

type Prefetcher struct {
	inFlightIds []uint32
	edgesFuture []*future[[]edge]
	locks       []sync.Mutex

	prefetchCache *caches.PrefetchCache[uint32, []edge]
	prefetchQueue *lists.CircularQueue[uint32]
	//This function will fetch all edges for a node.
	fetcher func(uint32) []edge
}

func NewPrefetcher(numThreads int, prefetchCacheSize int, fetcher func(uint32) []edge) *Prefetcher {
	pf := &Prefetcher{
		inFlightIds:   make([]uint32, numThreads),
		edgesFuture:   make([]*future[[]edge], numThreads),
		locks:         make([]sync.Mutex, numThreads),
		prefetchCache: caches.NewPrefetchCache[uint32, []edge](prefetchCacheSize),
		prefetchQueue: lists.NewCircularQueue[uint32](100),
		fetcher:       fetcher,
	}
	for i := 0; i < numThreads; i++ {
		go pf.prefetchRoutine(i)
	}
	return pf
}

func (pf *Prefetcher) write(result []uint32) {
	pf.prefetchQueue.Write(result)
}

func (pf *Prefetcher) prefetchRoutine(index int) {
	for {
		val := pf.prefetchQueue.Read()

		pf.locks[index].Lock()
		pf.inFlightIds[index] = val
		pf.edgesFuture[index] = newFuture[[]edge]()
		pf.locks[index].Unlock()

		resultEdges := pf.fetcher(val)

		pf.locks[index].Lock()
		pf.edgesFuture[index].put(resultEdges)
		pf.inFlightIds[index] = 0
		pf.edgesFuture[index] = nil
		pf.locks[index].Unlock()

		pf.prefetchCache.Put(val, resultEdges)
	}
}

func (pf *Prefetcher) getFromPrefetchCache(node uint32) ([]edge, bool) {
	return pf.prefetchCache.Get(node)
}

func (pf *Prefetcher) getFromInFlightQueue(node uint32) (*future[[]edge], bool) {
	for i := 0; i < len(pf.inFlightIds); i++ {
		pf.locks[i].Lock()
		if pf.inFlightIds[i] == node {
			res := pf.edgesFuture[i]
			pf.locks[i].Unlock()
			return res, true
		}
		pf.locks[i].Unlock()
	}
	return nil, false
}
