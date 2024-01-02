package graphaccess

import "testing"

func TestGraphReprEdges(t *testing.T) {
	repr := simpleCsrRepr{
		startNodeId: 0,
		indices:     []uint32{0, 2, 2, 4},
		edges:       []edge{{1, 3}, {2, 4}, {1, 5}, {1, 8}, {2, 5}, {2, 8}},
	}
	values := repr.getEdges(1, 2)
	assertLength(t, values, 0)
	values = repr.getEdges(2, 1)
	assertLength(t, values, 2)
	values = repr.getEdges(0, 2)
	assertLength(t, values, 1)
	values = repr.getEdges(3, 2)
	assertLength(t, values, 2)
}

func assertLength[K any](t *testing.T, slice []K, size int) {
	if len(slice) != size {
		t.Fatalf("Expected size %d, got %d", size, len(slice))
	}
}
