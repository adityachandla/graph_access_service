package graphaccess

import (
	"github.com/adityachandla/graph_access_service/caches"
	"github.com/adityachandla/graph_access_service/storage"
	"sync"
)

const NumFetchers = 5
const PrefetchQueueSize = 100

type PrefetchCsr struct {
	offsetCsr *OffsetCsr
	inFlight  [NumFetchers]inFlightNode
	queue     prefetchQueue
	cache     *caches.Lrfu[uint32, []edge]
}

type prefetchQueue struct {
	toFetchArray [PrefetchQueueSize]uint32
	head         int
	lock         sync.Mutex
}

type inFlightNode struct {
	nodeInFlight uint32
	lock         sync.Mutex
	edgesFuture  future[[]edge]
}

func NewPrefetchCsr(fetcher storage.Fetcher) *PrefetchCsr {
	return &PrefetchCsr{
		offsetCsr: NewOffsetCsr(fetcher),
	}
}
