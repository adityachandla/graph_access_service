package graphaccess

const SizeIntBytes = 4

type GraphAccess interface {
	StartQuery(Algo) int
	GetNeighbours(Request, int) []uint32
	EndQuery(int)
	GetStats() string
}

type Algo byte

const (
	BFS Algo = iota
	DFS
)

type Request struct {
	Node, Label uint32
	Direction   Direction
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
func getEdgesWithLabelBytes(edges edgeList, label uint32) []uint32 {
	low := 0
	high := edges.Len() - 1
	for low <= high {
		mid := low + (high-low)/2
		if edges.LabelAt(mid) <= label {
			low = mid + 1
		} else {
			high = mid - 1
		}
	}
	labelEnd := high
	low = 0
	high = edges.Len() - 1
	for low <= high {
		mid := low + (high-low)/2
		if edges.LabelAt(mid) < label {
			low = mid + 1
		} else {
			high = mid - 1
		}
	}
	labelStart := low
	res := make([]uint32, 0, labelEnd-labelStart+1)
	for i := labelStart; i <= labelEnd; i++ {
		res = append(res, edges.NodeAt(i))
	}
	return res
}

func getEdgesWithLabel(edges []edge, label uint32) []uint32 {
	low := 0
	high := len(edges) - 1
	for low <= high {
		mid := low + (high-low)/2
		if edges[mid].label <= label {
			low = mid + 1
		} else {
			high = mid - 1
		}
	}
	labelEnd := high
	low = 0
	high = len(edges) - 1
	for low <= high {
		mid := low + (high-low)/2
		if edges[mid].label < label {
			low = mid + 1
		} else {
			high = mid - 1
		}
	}
	labelStart := low
	res := make([]uint32, 0, labelEnd-labelStart+1)
	for i := labelStart; i <= labelEnd; i++ {
		res = append(res, edges[i].dest)
	}
	return res
}
