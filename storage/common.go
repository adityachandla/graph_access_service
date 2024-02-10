package storage

type Fetcher interface {
	Fetch(objectName string, bRange ByteRange) []byte
	ListFiles() []string
}

type ByteRange struct {
	start, end uint32
}

func BRangeStart(start uint32) ByteRange {
	return ByteRange{start: start, end: 0}
}

func BRange(start, end uint32) ByteRange {
	return ByteRange{start: start, end: end}
}
