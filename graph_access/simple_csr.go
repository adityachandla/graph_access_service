package graph_access

import (
	"fmt"
	"log"
	"slices"
	"sync"
	"unsafe"

	"github.com/adityachandla/graph_access_service/bin_util"
	"github.com/adityachandla/graph_access_service/caches"
	"github.com/adityachandla/graph_access_service/s3_util"
)

const LRU_SIZE_FILES = 5

type simpleCsrAccess struct {
	nodePaths []nodeRangePath
	lru       *caches.LRU[string, simpleCsrRepr]
}

type csrFetcher struct {
	s3Util *s3_util.S3Service
}

func (f csrFetcher) Fetch(objectName string) *simpleCsrRepr {
	fileBytes := f.s3Util.Fetch(objectName, s3_util.ByteRangeStart(0))
	start := bin_util.ByteToUint(fileBytes[:4])
	end := bin_util.ByteToUint(fileBytes[4:8])
	numValues := end - start + 1
	sizeOfIndices := SIZE_INT_BYTES * numValues
	nodeIndices := bin_util.ByteArrayToUintArray(fileBytes[8 : 8+sizeOfIndices])
	pairs := bin_util.ByteArrayToPairArray(fileBytes[8+sizeOfIndices:])
	//The memory layout of pair is same as edge so it is safe to
	//do a typecast.
	pairPtr := unsafe.Pointer(&pairs)
	return &simpleCsrRepr{
		startNodeId: start,
		indices:     nodeIndices,
		edges:       *(*[]edge)(pairPtr),
	}
}

type simpleCsrRepr struct {
	startNodeId uint32
	indices     []uint32
	edges       []edge
}

func (repr *simpleCsrRepr) getEdges(src, label uint32) []uint32 {
	index := src - repr.startNodeId
	edgeStart := int(repr.indices[index])
	//edgeEnd is exclusive
	var edgeEnd int
	if int(index) == len(repr.indices)-1 {
		edgeEnd = len(repr.edges)
	} else {
		edgeEnd = int(repr.indices[index+1])
	}
	edges := make([]uint32, 0)
	//TODO optimize with a binary search
	for i := edgeStart; i < edgeEnd; i++ {
		if repr.edges[i].label == label {
			edges = append(edges, repr.edges[i].dest)
		}
	}
	return edges
}

func (scsr *simpleCsrAccess) GetNeighbours(src, label uint32,
	incoming bool) ([]uint32, error) {
	if incoming {
		return []uint32{}, IncomingNotImplemented
	}
	objectName := scsr.getObjectWithNode(src)
	csrRepr := scsr.lru.Get(objectName)
	return csrRepr.getEdges(src, label), nil
}

func (scsr *simpleCsrAccess) getObjectWithNode(src uint32) string {
	start := 0
	end := len(scsr.nodePaths) - 1
	for start <= end {
		mid := (start + end) / 2
		if scsr.nodePaths[mid].contains(src) {
			return scsr.nodePaths[mid].objectName
		} else if scsr.nodePaths[mid].start > src {
			start = mid - 1
		} else {
			end = mid + 1
		}
	}
	panic(fmt.Errorf("%d not found in nodeRanges", src))
}

func InitializeSimpleCsrAccess(s3 *s3_util.S3Service) *simpleCsrAccess {
	objects := s3.GetFilesInBucket()
	//For each object, we need to fetch the start and end stored in that file.
	//Start and end will be the first 8 bytes of the file.
	nodePaths := make([]nodeRangePath, len(objects))
	bRange := s3_util.ByteRange(0, 8)
	var wg sync.WaitGroup
	for i := range objects {
		idx := i
		wg.Add(1)
		go func() {
			defer wg.Done()
			startEndBytes := s3.Fetch(objects[idx], bRange)
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
