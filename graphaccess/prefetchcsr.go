package graphaccess

import (
	"encoding/json"
	"github.com/adityachandla/graph_access_service/caches"
	"github.com/adityachandla/graph_access_service/storage"
	"sync/atomic"
)

const NumFetchers = 5

type PrefetchCsr struct {
	offsetCsr  *OffsetCsr
	prefetcher *Prefetcher
	cache      *caches.Lrfu[Request, []uint32]
	stats      PrefetchStats
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
	p := &PrefetchCsr{
		offsetCsr: NewOffsetCsr(fetcher),
		cache:     caches.NewLrfuCache[Request, []uint32](1000, 0.2),
	}
	p.prefetcher = NewPrefetcher(NumFetchers, 100, p.offsetCsr.fetchAllEdges)
	return p
}

func (p *PrefetchCsr) GetNeighbours(req Request) []uint32 {
	response := p.fetchResponse(req)
	if !p.cache.Present(req) {
		p.cache.Put(req, response)
	}
	p.prefetcher.write(response)
	return response
}

func (p *PrefetchCsr) GetStats() string {
	return p.stats.convertToString()
}

func (p *PrefetchCsr) fetchResponse(req Request) []uint32 {
	//Check the LRFU cache
	response, found := p.cache.Get(req)
	if found {
		p.stats.CacheHits.Add(1)
		return response
	}
	//Then check the Prefetcher cache
	edges, found := p.prefetcher.getFromPrefetchCache(req.Node)
	if found {
		p.stats.PrefetcherHits.Add(1)
		return filterResponse(req, edges, p.offsetCsr.offsets)
	}
	//Then check the in-flight queue
	edgesFuture, found := p.prefetcher.getFromInFlightQueue(req.Node)
	if found {
		p.stats.InFlightHits.Add(1)
		return filterResponse(req, edgesFuture.get(), p.offsetCsr.offsets)
	}
	//Fetch from S3
	p.stats.S3Fetches.Add(1)
	return p.offsetCsr.GetNeighbours(req)
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
