package graphaccess

import (
	"fmt"
	"github.com/adityachandla/graph_access_service/bin_util"
	"github.com/adityachandla/graph_access_service/storage"
	"slices"
	"sync"
	"unsafe"
)

type OffsetCsr struct {
	offsets fileOffsets
	fetcher storage.Fetcher
}

func NewOffsetCsr(fetcher storage.Fetcher) *OffsetCsr {
	files := fetcher.ListFiles()
	offsets := make(fileOffsets, 0, len(files))
	fileOffsetChannel := make(chan *fileOffset)
	wg := sync.WaitGroup{}
	for _, f := range files {
		wg.Add(1)
		go func(filename string) {
			defer wg.Done()
			fetchFileOffset(filename, fetcher, fileOffsetChannel)
		}(f)
	}
	go func() {
		wg.Wait()
		close(fileOffsetChannel)
	}()
	for offset := range fileOffsetChannel {
		offsets = append(offsets, offset)
	}
	slices.SortFunc(offsets, func(a, b *fileOffset) int {
		if a.nodeRange.start > b.nodeRange.start {
			return 1
		}
		return -1
	})
	return &OffsetCsr{offsets, fetcher}
}

func fetchFileOffset(filename string, fetcher storage.Fetcher, outputChannel chan<- *fileOffset) {
	startEndBytes := fetcher.Fetch(filename, storage.BRange(0, 7))
	start := bin_util.ByteToUint(startEndBytes[0:4])
	end := bin_util.ByteToUint(startEndBytes[4:])
	nodeRange := nodeRangePath{
		start:      start,
		end:        end,
		objectName: filename,
	}
	sizeOfOffsets := 2 * SizeIntBytes * (end - start + 1)
	offsetBytes := fetcher.Fetch(filename, storage.BRange(8, 8+sizeOfOffsets-1))
	offsetPairs := bin_util.ByteArrayToPairArray(offsetBytes)
	offsetStruct := &fileOffset{
		nodeRange: nodeRange,
		offsetArr: *(*[]nodeOffset)(unsafe.Pointer(&offsetPairs)),
	}
	outputChannel <- offsetStruct
}

func (csr *OffsetCsr) StartQuery(Algo) int {
	//This implementation does not do anything about the queries.
	return 1
}

func (csr *OffsetCsr) GetNeighbours(req Request, _ int) []uint32 {
	file := csr.offsets.find(req.Node)
	offset, numOut := file.fetchOffset(req)

	resultBytes := csr.fetcher.Fetch(file.nodeRange.objectName, offset)
	resultPairs := bin_util.ByteArrayToPairArray(resultBytes)
	resultEdges := *(*[]edge)(unsafe.Pointer(&resultPairs))

	if req.Direction != BOTH {
		return getEdgesWithLabel(resultEdges, req.Label)
	}
	filtered := getEdgesWithLabel(resultEdges[:numOut], req.Label)
	return append(filtered, getEdgesWithLabel(resultEdges[numOut:], req.Label)...)
}

func (csr *OffsetCsr) EndQuery(int) {
}

func (csr *OffsetCsr) fetchAllEdges(node uint32) []edge {
	file := csr.offsets.find(node)
	byteRange := file.fetchOffsetAllEdges(node)

	resultBytes := csr.fetcher.Fetch(file.nodeRange.objectName, byteRange)
	resultPairs := bin_util.ByteArrayToPairArray(resultBytes)
	return *(*[]edge)(unsafe.Pointer(&resultPairs))
}

func (csr *OffsetCsr) GetStats() string {
	return ""
}

type fileOffsets []*fileOffset

func (fo fileOffsets) find(node uint32) *fileOffset {
	low := 0
	high := len(fo) - 1
	for low <= high {
		mid := low + ((high - low) / 2)
		if fo[mid].contains(node) {
			return fo[mid]
		} else if fo[mid].nodeRange.start > node {
			high = mid - 1
		} else {
			low = mid + 1
		}
	}
	panic(fmt.Errorf("Node %d not found in fileOffsets\n", node))
}

type fileOffset struct {
	//nodeRange stores the filename along with the start and end node information.
	nodeRange nodeRangePath
	offsetArr []nodeOffset
}

func (offset *fileOffset) contains(node uint32) bool {
	return node >= offset.nodeRange.start && node <= offset.nodeRange.end
}

func (offset *fileOffset) fetchOffset(req Request) (storage.ByteRange, uint32) {
	idx := req.Node - offset.nodeRange.start
	if req.Direction == OUTGOING {
		start := offset.offsetArr[idx].outgoing
		numOut := (offset.offsetArr[idx].incoming - start) / (2 * SizeIntBytes)
		return storage.BRange(start, offset.offsetArr[idx].incoming-1), numOut
	} else if req.Direction == INCOMING {
		start := offset.offsetArr[idx].incoming
		if int(idx) < len(offset.offsetArr)-1 {
			return storage.BRange(start, offset.offsetArr[idx+1].outgoing-1), 0
		} else {
			return storage.BRangeStart(start), 0
		}
	}
	//Both incoming and outgoing
	start := offset.offsetArr[idx].outgoing
	numOut := (offset.offsetArr[idx].incoming - start) / (2 * SizeIntBytes)
	if int(idx) < len(offset.offsetArr)-1 {
		return storage.BRange(start, offset.offsetArr[idx+1].outgoing-1), numOut
	} else {
		return storage.BRangeStart(start), numOut
	}
}
func (offset *fileOffset) fetchOffsetAllEdges(node uint32) storage.ByteRange {
	idx := node - offset.nodeRange.start
	start := offset.offsetArr[idx].outgoing
	if int(idx) < len(offset.offsetArr)-1 {
		return storage.BRange(start, offset.offsetArr[idx+1].outgoing-1)
	} else {
		return storage.BRangeStart(start)
	}
}

type nodeOffset struct {
	outgoing, incoming uint32
}
