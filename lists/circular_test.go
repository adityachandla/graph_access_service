package lists_test

import (
	"github.com/adityachandla/graph_access_service/lists"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNormalAddition(t *testing.T) {
	cq := lists.NewCircularQueue[int](4)
	_, ok := cq.Read()
	assert.False(t, ok)
	cq.Write([]int{2, 3})

	v, ok := cq.Read()
	assert.True(t, ok)
	assert.Equal(t, 3, v)

	v, ok = cq.Read()
	assert.True(t, ok)
	assert.Equal(t, 2, v)
}

func TestOverwrite(t *testing.T) {
	cq := lists.NewCircularQueue[int](4)
	cq.Write([]int{4, 3, 2, 1})
	cq.Write([]int{5, 6})
	readRes := []int{3, 4, 6, 5}
	for _, v := range readRes {
		value, ok := cq.Read()
		assert.True(t, ok)
		assert.Equal(t, v, value)
	}
}
