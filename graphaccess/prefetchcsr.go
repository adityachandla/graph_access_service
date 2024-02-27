package graphaccess

import (
	"encoding/json"
	"github.com/adityachandla/graph_access_service/caches"
	"github.com/adityachandla/graph_access_service/storage"
	"sync"
	"sync/atomic"
)

type PrefetchCsr struct {
	offsetCsr      *OffsetCsr
	mapLock        sync.Mutex
	prefetchers    map[int]*Prefetcher
	cache          *caches.Lrfu[Request, []uint32]
	queryIdCounter int
	stats          PrefetchStats
}

type PrefetchStats struct {
	CacheHits      atomic.Uint32
	PrefetcherHits atomic.Uint32
	InFlightHits   atomic.Uint32
	S3Fetches      atomic.Uint32
}

func (s *PrefetchStats) convertToString() string {
	res := make(map[string]uint32, 4)
	res["cacheHits"] = s.CacheHits.Load()
	res["prefetcherHits"] = s.PrefetcherHits.Load()
	res["inFlightHits"] = s.InFlightHits.Load()
	res["S3Fetches"] = s.S3Fetches.Load()
	resultBytes, err := json.Marshal(res)
	if err != nil {
		panic(err)
	}
	return string(resultBytes)
}

func NewPrefetchCsr(fetcher storage.Fetcher) *PrefetchCsr {
	return &PrefetchCsr{
		offsetCsr:      NewOffsetCsr(fetcher),
		cache:          caches.NewLrfuCache[Request, []uint32](1000, 0.2),
		queryIdCounter: 1,
		prefetchers:    make(map[int]*Prefetcher),
	}
}

func (p *PrefetchCsr) StartQuery(algo Algo) int {
	val := p.queryIdCounter
	p.queryIdCounter++
	p.mapLock.Lock()
	p.prefetchers[val] = NewPrefetcher(algo, p.offsetCsr.fetchAllEdges)
	p.mapLock.Unlock()
	return val
}

func (p *PrefetchCsr) GetNeighbours(req Request, queryId int) []uint32 {
	response := p.fetchResponse(req, queryId)
	if !p.cache.Present(req) {
		p.cache.Put(req, response)
	}
	p.prefetchers[queryId].write(response)
	return response
}

func (p *PrefetchCsr) EndQuery(id int) {
	p.mapLock.Lock()
	defer p.mapLock.Unlock()
	if pf, ok := p.prefetchers[id]; ok {
		pf.StopPrefetcher()
		delete(p.prefetchers, id)
	}
}

func (p *PrefetchCsr) GetStats() string {
	return p.stats.convertToString()
}

func (p *PrefetchCsr) fetchResponse(req Request, queryId int) []uint32 {
	//Check the LRFU cache
	response, found := p.cache.Get(req)
	if found {
		p.stats.CacheHits.Add(1)
		return response
	}
	//Then check the Prefetcher cache
	edges, found := p.prefetchers[queryId].getFromPrefetchCache(req.Node)
	if found {
		p.stats.PrefetcherHits.Add(1)
		return filterResponse(req, edges, p.offsetCsr.offsets)
	}
	//Then check the in-flight queue
	edgesFuture, found := p.prefetchers[queryId].getFromInFlightQueue(req.Node)
	if found {
		p.stats.InFlightHits.Add(1)
		return filterResponse(req, edgesFuture.get(), p.offsetCsr.offsets)
	}
	//Fetch from S3
	p.stats.S3Fetches.Add(1)
	return p.offsetCsr.GetNeighbours(req, queryId)
}

func filterResponse(req Request, edges edgeList, offsets fileOffsets) []uint32 {
	_, numOutgoing := offsets.find(req.Node).fetchOffset(req)
	if req.Direction == OUTGOING {
		return getEdgesWithLabelBytes(edges.SliceEnd(int(numOutgoing)), req.Label)
	} else if req.Direction == INCOMING {
		return getEdgesWithLabelBytes(edges.SliceStart(int(numOutgoing)), req.Label)
	}
	//Both
	outgoing := getEdgesWithLabelBytes(edges.SliceEnd(int(numOutgoing)), req.Label)
	return append(outgoing, getEdgesWithLabelBytes(edges.SliceStart(int(numOutgoing)), req.Label)...)
}
