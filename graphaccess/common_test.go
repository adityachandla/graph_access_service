package graphaccess

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func edgesToBytes(edges []edge) edgeList {
	l := make([]byte, 0, len(edges)*8)
	for _, e := range edges {
		l = append(l, byte(e.label), 0, 0, 0)
		l = append(l, byte(e.dest), 0, 0, 0)
	}
	return l
}

func TestEdgesMiddle(t *testing.T) {
	edges := []edge{{1, 2}, {1, 4}, {2, 3}, {2, 4}, {2, 5}, {3, 1}, {3, 3}}
	edgesBytes := edgesToBytes(edges)
	assert.Equal(t, []uint32{3, 4, 5}, getEdgesWithLabelBytes(edgesBytes, 2))
	assert.Equal(t, []uint32{3, 4}, getEdgesWithLabelBytes(edgesBytes.SliceEnd(4), 2))
	assert.Equal(t, []uint32{3}, getEdgesWithLabelBytes(edgesBytes.SliceEnd(3), 2))
	assert.Equal(t, []uint32{5}, getEdgesWithLabelBytes(edgesBytes.SliceStart(4), 2))
	assert.Equal(t, []uint32{}, getEdgesWithLabelBytes([]byte{}, 2))
}
