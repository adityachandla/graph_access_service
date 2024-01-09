package storage

type Fetcher interface {
	Fetch(objectName string, bRange byteRange) []byte
	ListFiles() []string
}

type byteRange struct {
	start, end uint32
}

func ByteRangeStart(start uint32) byteRange {
	return byteRange{start: start, end: 0}
}

func ByteRange(start, end uint32) byteRange {
	return byteRange{start: start, end: end}
}
