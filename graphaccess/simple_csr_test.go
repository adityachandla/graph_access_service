package graphaccess

import (
	"testing"

	pb "github.com/adityachandla/graph_access_service/generated"
)

func TestGraphReprEdges(t *testing.T) {
	repr := simpleCsrRepr{
		startNodeId: 0,
		indices:     []nodeIndex{{0, 2}, {2, 2}, {2, 4}, {4, 6}},
		edges:       []edge{{1, 3}, {2, 4}, {1, 5}, {1, 8}, {2, 5}, {2, 8}},
	}
	values := repr.getEdges(&pb.AccessRequest{NodeId: 1, Label: 2, Direction: pb.AccessRequest_OUTGOING})
	assertLength(t, values, 0)
	values = repr.getEdges(&pb.AccessRequest{NodeId: 2, Label: 1, Direction: pb.AccessRequest_OUTGOING})
	assertLength(t, values, 2)
	values = repr.getEdges(&pb.AccessRequest{NodeId: 0, Label: 2, Direction: pb.AccessRequest_OUTGOING})
	assertLength(t, values, 1)
	values = repr.getEdges(&pb.AccessRequest{NodeId: 3, Label: 2, Direction: pb.AccessRequest_OUTGOING})
	assertLength(t, values, 2)
}

func assertLength[K any](t *testing.T, slice []K, size int) {
	if len(slice) != size {
		t.Fatalf("Expected size %d, got %d", size, len(slice))
	}
}
