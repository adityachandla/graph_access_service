package graphaccess

import (
	"fmt"
	"log"
	"slices"
	"sync"
	"unsafe"

	"github.com/adityachandla/graph_access_service/bin_util"
	"github.com/adityachandla/graph_access_service/caches"
	"github.com/adityachandla/graph_access_service/storage"
)

const LruSizeFiles = 7

type csr struct {
	nodePaths []nodeRangePath
	lru       *caches.LRU[string, csrRepr]
	fetcher   storage.Fetcher
}

type csrRepr struct {
	startNodeId uint32
	indices     []nodeIndex
	edges       []edge
}

type nodeIndex struct {
	outgoing, incoming uint32
}

func InitializeSimpleCsrAccess(fetcher storage.Fetcher) *csr {
	objects := fetcher.ListFiles()
	//For each object, we need to fetch the start and end stored in that file.
	//Start and end will be the first 8 bytes of the file.
	nodePaths := make([]nodeRangePath, len(objects))
	bRange := storage.ByteRange(0, 7)
	var wg sync.WaitGroup
	for i := range objects {
		idx := i
		wg.Add(1)
		go func() {
			defer wg.Done()
			startEndBytes := fetcher.Fetch(objects[idx], bRange)
			nodePaths[idx].start = bin_util.ByteToUint(startEndBytes[:4])
			nodePaths[idx].end = bin_util.ByteToUint(startEndBytes[4:])
			nodePaths[idx].objectName = objects[idx]
		}()
	}
	wg.Wait()
	log.Println("Initialized simple csr")
	slices.SortFunc(nodePaths, nodeCmp)
	return &csr{
		nodePaths: nodePaths,
		lru:       caches.NewLRU[string, csrRepr](LruSizeFiles),
		fetcher:   fetcher,
	}
}

func (scsr *csr) Fetch(objectName string) csrRepr {
	log.Printf("Fetching %s\n", objectName)
	fileBytes := scsr.fetcher.Fetch(objectName, storage.ByteRangeStart(0))
	start := bin_util.ByteToUint(fileBytes[:4])
	end := bin_util.ByteToUint(fileBytes[4:8])
	numValues := end - start + 1
	sizeOfIndices := 2 * SizeIntBytes * numValues
	nodeIndices := bin_util.ByteArrayToPairArray(fileBytes[8 : 8+sizeOfIndices])
	pairs := bin_util.ByteArrayToPairArray(fileBytes[8+sizeOfIndices:])
	//The memory layout of pair is same as edge so, it is safe to
	//do a direct typecast.
	pairPtr := unsafe.Pointer(&pairs)
	nodeIndexPtr := unsafe.Pointer(&nodeIndices)
	return csrRepr{
		startNodeId: start,
		indices:     *(*[]nodeIndex)(nodeIndexPtr),
		edges:       *(*[]edge)(pairPtr),
	}
}

func (repr *csrRepr) getEdges(req Request) []uint32 {
	//Incoming or outgoing
	if req.Direction != BOTH {
		incoming := req.Direction == INCOMING
		edgeStart := repr.getStartEdgeIndex(req.Node, incoming)
		edgeEnd := repr.getEndEdgeIndex(req.Node, incoming)
		return getEdgesWithLabel(repr.edges[edgeStart:edgeEnd], req.Label)
	}

	//Both
	outStart := repr.getStartEdgeIndex(req.Node, false)
	outEnd := repr.getEndEdgeIndex(req.Node, false)
	out := getEdgesWithLabel(repr.edges[outStart:outEnd], req.Label)

	inStart := repr.getStartEdgeIndex(req.Node, true)
	inEnd := repr.getEndEdgeIndex(req.Node, true)
	in := getEdgesWithLabel(repr.edges[inStart:inEnd], req.Label)
	out = append(out, in...)
	return out
}

func (repr *csrRepr) getStartEdgeIndex(nodeId uint32, incoming bool) uint32 {
	index := nodeId - repr.startNodeId
	if incoming {
		return repr.indices[index].incoming
	}
	return repr.indices[index].outgoing
}

func (repr *csrRepr) getEndEdgeIndex(nodeId uint32, incoming bool) uint32 {
	index := nodeId - repr.startNodeId
	//Special case for end of file.
	if int(index) == len(repr.indices)-1 && incoming {
		return uint32(len(repr.edges))
	}
	if incoming {
		// For incoming, we look at the next index.
		return repr.indices[index+1].outgoing
	}
	// For outgoing, we look at the start of incoming edges.
	return repr.indices[index].incoming
}

func (scsr *csr) GetNeighbours(req Request) ([]uint32, error) {
	objectName := scsr.getObjectWithNode(req.Node)
	csrRepr, found := scsr.lru.Get(objectName)
	if !found {
		csrRepr = scsr.Fetch(objectName)
		scsr.lru.Put(objectName, csrRepr)
	}
	return csrRepr.getEdges(req), nil
}

func (scsr *csr) getObjectWithNode(src uint32) string {
	start := 0
	end := len(scsr.nodePaths) - 1
	for start <= end {
		mid := (start + end) / 2
		if scsr.nodePaths[mid].contains(src) {
			return scsr.nodePaths[mid].objectName
		} else if scsr.nodePaths[mid].start > src {
			end = mid - 1
		} else {
			start = mid + 1
		}
	}
	panic(fmt.Errorf("%d not found in nodeRanges", src))
}

func nodeCmp(one, two nodeRangePath) int {
	if one.start > two.start {
		return 1
	}
	if one.start < two.start {
		return -1
	}
	return 0
}
