package graphaccess

import "fmt"

var IncomingNotImplemented error = fmt.Errorf("Incoming edge query not implemented")

const SIZE_INT_BYTES = 4

type GraphAccess interface {
	GetNeighbours(src, label uint32, incoming bool) ([]uint32, error)
}

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
