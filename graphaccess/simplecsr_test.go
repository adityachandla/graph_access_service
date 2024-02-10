package graphaccess

import (
	"testing"
)

func TestGraphReprEdges(t *testing.T) {
	repr := csrRepr{
		startNodeId: 0,
		indices:     []nodeIndex{{0, 2}, {2, 2}, {2, 4}, {4, 6}},
		edges:       []edge{{1, 3}, {2, 4}, {1, 5}, {1, 8}, {2, 5}, {2, 8}},
	}
	values := repr.getEdges(Request{Node: 1, Label: 2, Direction: OUTGOING})
	assertLength(t, values, 0)
	values = repr.getEdges(Request{Node: 2, Label: 1, Direction: OUTGOING})
	assertLength(t, values, 2)
	values = repr.getEdges(Request{Node: 0, Label: 2, Direction: OUTGOING})
	assertLength(t, values, 1)
	values = repr.getEdges(Request{Node: 3, Label: 2, Direction: OUTGOING})
	assertLength(t, values, 2)
}

func assertLength[K any](t *testing.T, slice []K, size int) {
	if len(slice) != size {
		t.Fatalf("Expected size %d, got %d", size, len(slice))
	}
}
