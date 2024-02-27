package graphaccess

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestEdgeList(t *testing.T) {
	e := edgeList([]byte{1, 0, 0, 0, 2, 0, 0, 0, 3, 0, 0, 0, 4, 0, 0, 0})
	assert.Equal(t, 2, e.Len())

	assert.Equal(t, uint32(1), e.LabelAt(0))
	assert.Equal(t, uint32(2), e.NodeAt(0))

	assert.Equal(t, uint32(3), e.LabelAt(1))
	assert.Equal(t, uint32(4), e.NodeAt(1))
}
