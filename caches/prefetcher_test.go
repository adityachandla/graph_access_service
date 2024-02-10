package caches

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestRemoveOnRead(t *testing.T) {
	pc := NewPrefetchCache[int, int](2)
	pc.Put(1, 101)
	res, found := pc.Get(1)
	assert.True(t, found)
	assert.Equal(t, 101, res)
	assert.Equal(t, 0, pc.Len())
}

func TestEviction(t *testing.T) {
	pc := NewPrefetchCache[int, int](2)
	pc.Put(1, 101)
	pc.Put(2, 102)
	pc.Put(3, 103)
	_, found := pc.Get(1)
	assert.False(t, found)
	assert.Equal(t, 2, pc.Len())
}
