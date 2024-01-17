package graphaccess

import (
	"fmt"
	"log"
	"slices"
	"sync"
	"unsafe"

	"github.com/adityachandla/graph_access_service/bin_util"
	"github.com/adityachandla/graph_access_service/caches"
	pb "github.com/adityachandla/graph_access_service/generated"
	"github.com/adityachandla/graph_access_service/storage"
)

const LRU_SIZE_FILES = 7

type simpleCsrAccess struct {
	nodePaths []nodeRangePath
	lru       *caches.LRU[string, simpleCsrRepr]
}

// This type implements caches.Fetcher interface.
// It is passed to the LRU cache so that the cache
// can control when an object is fetched. This struct
// simply provides the fetching logic.
type csrFetcher struct {
	fetcher storage.Fetcher
}

func (f csrFetcher) Fetch(objectName string) simpleCsrRepr {
	log.Printf("Fetching %s\n", objectName)
	fileBytes := f.fetcher.Fetch(objectName, storage.ByteRangeStart(0))
	start := bin_util.ByteToUint(fileBytes[:4])
	end := bin_util.ByteToUint(fileBytes[4:8])
	numValues := end - start + 1
	sizeOfIndices := 2 * SIZE_INT_BYTES * numValues
	nodeIndices := bin_util.ByteArrayToPairArray(fileBytes[8 : 8+sizeOfIndices])
	pairs := bin_util.ByteArrayToPairArray(fileBytes[8+sizeOfIndices:])
	//The memory layout of pair is same as edge so it is safe to
	//do a typecast.
	pairPtr := unsafe.Pointer(&pairs)
	nodeIndexPtr := unsafe.Pointer(&nodeIndices)
	return simpleCsrRepr{
		startNodeId: start,
		indices:     *(*[]nodeIndex)(nodeIndexPtr),
		edges:       *(*[]edge)(pairPtr),
	}
}

type simpleCsrRepr struct {
	startNodeId uint32
	indices     []nodeIndex
	edges       []edge
}

type nodeIndex struct {
	outgoing, incoming uint32
}

func (repr *simpleCsrRepr) getEdges(req *pb.AccessRequest) []uint32 {
	edgeStart := repr.getStartEdgeIndex(req)
	edgeEnd := repr.getEndEdgeIndex(req)
	return getEdgesWithLabel(repr.edges[edgeStart:edgeEnd], req.Label)
}

func (repr *simpleCsrRepr) getStartEdgeIndex(req *pb.AccessRequest) uint32 {
	index := req.NodeId - repr.startNodeId
	if req.Direction == pb.AccessRequest_INCOMING {
		return repr.indices[index].incoming
	}
	return repr.indices[index].outgoing
}

func (repr *simpleCsrRepr) getEndEdgeIndex(req *pb.AccessRequest) uint32 {
	index := req.NodeId - repr.startNodeId
	if int(index) == len(repr.indices)-1 && req.Direction != pb.AccessRequest_OUTGOING {
		return uint32(len(repr.edges))
	}
	if req.Direction != pb.AccessRequest_OUTGOING {
		// For incoming or both, we look at the next index.
		return repr.indices[index+1].outgoing
	}
	// For outgoing, we look at the start of incoming edges.
	return repr.indices[index].incoming
}

func (scsr *simpleCsrAccess) GetNeighbours(req *pb.AccessRequest) ([]uint32, error) {
	objectName := scsr.getObjectWithNode(req.NodeId)
	csrRepr := scsr.lru.Get(objectName)
	return csrRepr.getEdges(req), nil
}

func (scsr *simpleCsrAccess) getObjectWithNode(src uint32) string {
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

func InitializeSimpleCsrAccess(fetcher storage.Fetcher) *simpleCsrAccess {
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
	return &simpleCsrAccess{
		nodePaths: nodePaths,
		lru:       caches.NewLRU(&csrFetcher{fetcher}, LRU_SIZE_FILES),
	}
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
