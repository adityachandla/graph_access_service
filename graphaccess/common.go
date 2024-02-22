package graphaccess

const SizeIntBytes = 4

type GraphAccess interface {
	StartQuery() int
	GetNeighbours(Request) []uint32
	EndQuery(int)
	GetStats() string
}

type Request struct {
	Node, Label uint32
	Direction   Direction
	QueryId     int
}

type Direction byte

const (
	INCOMING Direction = iota
	OUTGOING
	BOTH
)

// Stores the key of the file that
// stores nodes starting from `start`
// to `end` inclusive.
type nodeRangePath struct {
	start, end uint32
	objectName string
}

func (nr *nodeRangePath) contains(value uint32) bool {
	return nr.start <= value && nr.end >= value
}

type edge struct {
	label, dest uint32
}

// This function gets all the destinations with a particular label in the
// slice.
func getEdgesWithLabel(edgeList []edge, label uint32) []uint32 {
	low := 0
	high := len(edgeList) - 1
	for low <= high {
		mid := low + (high-low)/2
		if edgeList[mid].label <= label {
			low = mid + 1
		} else {
			high = mid - 1
		}
	}
	labelEnd := high
	low = 0
	high = len(edgeList) - 1
	for low <= high {
		mid := low + (high-low)/2
		if edgeList[mid].label < label {
			low = mid + 1
		} else {
			high = mid - 1
		}
	}
	labelStart := low
	res := make([]uint32, 0, labelEnd-labelStart+1)
	for i := labelStart; i <= labelEnd; i++ {
		res = append(res, edgeList[i].dest)
	}
	return res
}
