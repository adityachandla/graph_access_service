package graphaccess

import (
	"github.com/adityachandla/graph_access_service/caches"
	"github.com/adityachandla/graph_access_service/lists"
	"sync"
)

const NumPrefetchers = 5
const CacheSize = 100

type Prefetcher struct {
	inFlightIds []uint32
	edgesFuture []*future[[]edge]
	locks       []sync.Mutex

	prefetchCache *caches.PrefetchCache[uint32, []edge]
	prefetchQueue lists.Queue[uint32]
	//This function will fetch all edges for a node.
	fetcher     func(uint32) []edge
	quitChannel chan struct{}
}

func NewPrefetcher(algorithm Algo, fetcher func(uint32) []edge) *Prefetcher {
	pf := &Prefetcher{
		inFlightIds:   make([]uint32, NumPrefetchers),
		edgesFuture:   make([]*future[[]edge], NumPrefetchers),
		locks:         make([]sync.Mutex, NumPrefetchers),
		prefetchCache: caches.NewPrefetchCache[uint32, []edge](CacheSize),
		fetcher:       fetcher,
		quitChannel:   make(chan struct{}),
	}
	if algorithm == DFS {
		pf.prefetchQueue = lists.NewDFSQueue[uint32](100)
	} else {
		pf.prefetchQueue = lists.NewBFSQueue[uint32](100)
	}
	for i := 0; i < NumPrefetchers; i++ {
		go pf.prefetchRoutine(i)
	}
	return pf
}

func (pf *Prefetcher) StopPrefetcher() {
	pf.prefetchQueue.Delete()
}

func (pf *Prefetcher) write(result []uint32) {
	pf.prefetchQueue.WriteAll(result)
}

func (pf *Prefetcher) prefetchRoutine(index int) {
	for {
		val, deleted := pf.prefetchQueue.Read()
		if deleted {
			return
		}

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
