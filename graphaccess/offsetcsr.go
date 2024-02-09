package graphaccess

import (
	"github.com/adityachandla/graph_access_service/bin_util"
	"github.com/adityachandla/graph_access_service/storage"
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
	for offset := range fileOffsetChannel {
		offsets = append(offsets, offset)
	}
	return &OffsetCsr{offsets, fetcher}
}

func fetchFileOffset(filename string, fetcher storage.Fetcher, outputChannel chan<- *fileOffset) {
	startEndBytes := fetcher.Fetch(filename, storage.ByteRange(0, 8))
	start := bin_util.ByteToUint(startEndBytes[0:4])
	end := bin_util.ByteToUint(startEndBytes[4:])
	nodeRange := nodeRangePath{
		start:      start,
		end:        end,
		objectName: filename,
	}
	// TODO check if the size for last file needs to be greater.
	sizeOfOffsets := 2 * SizeIntBytes * (end - start)
	offsetBytes := fetcher.Fetch(filename, storage.ByteRange(8, 8+sizeOfOffsets))
	offsetPairs := bin_util.ByteArrayToPairArray(offsetBytes)
	offsetStruct := &fileOffset{
		nodeRange: nodeRange,
		offsetArr: *(*[]nodeOffset)(unsafe.Pointer(&offsetPairs)),
	}
	outputChannel <- offsetStruct
}

func (csr *OffsetCsr) GetNeighbours(req Request) (uint32, error) {
	//If incoming, fetch incoming, if outgoing, fetch outgoing and if both, then
	//fetch both.
}

type fileOffsets []*fileOffset

type fileOffset struct {
	//nodeRange stores the filename along with the start and end node information.
	nodeRange nodeRangePath
	offsetArr []nodeOffset
}

type nodeOffset struct {
	outgoing, incoming int
}
