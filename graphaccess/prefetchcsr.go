package graphaccess

import (
	"github.com/adityachandla/graph_access_service/bin_util"
	"github.com/adityachandla/graph_access_service/caches"
	"github.com/adityachandla/graph_access_service/lists"
	"github.com/adityachandla/graph_access_service/storage"
	"sync"
	"unsafe"
)

const NumFetchers = 5

type PrefetchCsr struct {
	offsetCsr     *OffsetCsr
	prefetchers   [NumFetchers]Prefetcher
	prefetchQueue *lists.CircularQueue[uint32]
	cache         *caches.Lrfu[Request, []uint32]
	prefetchCache *caches.PrefetchCache[uint32, []edge]
}

type Prefetcher struct {
	sync.Mutex
	nodeInFlight  uint32
	prefetchCache *caches.PrefetchCache[uint32, []edge]
	prefetchQueue *lists.CircularQueue[uint32]
	offsetCsr     *OffsetCsr
	edgesFuture   *future[[]edge]
}

func NewPrefetchCsr(fetcher storage.Fetcher) *PrefetchCsr {
	//TODO parameterize the constants.
	prefetcher := &PrefetchCsr{
		offsetCsr:     NewOffsetCsr(fetcher),
		prefetchQueue: lists.NewCircularQueue[uint32](100),
		prefetchCache: caches.NewPrefetchCache[uint32, []edge](100),
		cache:         caches.NewLrfuCache[Request, []uint32](1000, 0.2),
	}
	//Dispatch the go routines.
	for i := 0; i < len(prefetcher.prefetchers); i++ {
		prefetcher.prefetchers[i].prefetchCache = prefetcher.prefetchCache
		prefetcher.prefetchers[i].offsetCsr = prefetcher.offsetCsr
		prefetcher.prefetchers[i].prefetchQueue = prefetcher.prefetchQueue
		go prefetcher.prefetchers[i].prefetchRoutine()
	}
	return prefetcher
}

func (node *Prefetcher) prefetchRoutine() {
	for {
		val, ok := node.prefetchQueue.Read()
		if !ok {
			continue
		}
		node.nodeInFlight = val
		node.edgesFuture = newFuture[[]edge]()
		file := node.offsetCsr.offsets.find(val)
		byteRange := file.fetchOffsetAllEdges(val)

		resultBytes := node.offsetCsr.fetcher.Fetch(file.nodeRange.objectName, byteRange)
		resultPairs := bin_util.ByteArrayToPairArray(resultBytes)
		resultEdges := *(*[]edge)(unsafe.Pointer(&resultPairs))
		node.edgesFuture.put(resultEdges)
		node.prefetchCache.Put(val, resultEdges)
		node.nodeInFlight = 0
		node.edgesFuture = nil
	}
}

func (prefetcher *PrefetchCsr) GetNeighbours(req Request) []uint32 {
	response := prefetcher.fetchResponse(req)
	if !prefetcher.cache.Present(req) {
		prefetcher.cache.Put(req, response)
	}
	prefetcher.prefetchQueue.Write(response)
	return response
}

func (prefetcher *PrefetchCsr) fetchResponse(req Request) []uint32 {
	//Check the LRFU cache
	response, found := prefetcher.cache.Get(req)
	if found {
		return response
	}
	//Then check the Prefetcher cache
	edges, found := prefetcher.prefetchCache.Get(req.Node)
	if found {
		return filterResponse(req, edges, prefetcher.offsetCsr.offsets)
	}
	//Then check the in-flight queue
	edgesFuture, found := prefetcher.checkInFlight(req.Node)
	if found {
		return filterResponse(req, edgesFuture.get(), prefetcher.offsetCsr.offsets)
	}
	//Fetch from S3
	return prefetcher.offsetCsr.GetNeighbours(req)
}

func (prefetcher *PrefetchCsr) checkInFlight(node uint32) (*future[[]edge], bool) {
	for i := 0; i < NumFetchers; i++ {
		prefetcher.prefetchers[i].Lock()
		if prefetcher.prefetchers[i].nodeInFlight == node {
			edgesFuture := prefetcher.prefetchers[i].edgesFuture
			prefetcher.prefetchers[i].Unlock()
			return edgesFuture, true
		}
		prefetcher.prefetchers[i].Unlock()
	}
	return nil, false
}

func filterResponse(req Request, edges []edge, offsets fileOffsets) []uint32 {
	_, numOutgoing := offsets.find(req.Node).fetchOffset(req)
	if req.Direction == OUTGOING {
		return getEdgesWithLabel(edges[:numOutgoing], req.Label)
	} else if req.Direction == INCOMING {
		return getEdgesWithLabel(edges[numOutgoing:], req.Label)
	}
	outgoing := getEdgesWithLabel(edges[:numOutgoing], req.Label)
	return append(outgoing, getEdgesWithLabel(edges[numOutgoing:], req.Label)...)
}
