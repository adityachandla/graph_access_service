package caches

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestLinkedListStack(t *testing.T) {
	ll := newLinkedList[int, int]()
	ll.addToFront(2, 102)
	ll.addToFront(3, 103)
	v, err := ll.popFront()
	assert.Nil(t, err)
	assert.Equal(t, v.key, 3)
	assert.Equal(t, v.value, 103)
}

func TestLinkedListQueue(t *testing.T) {
	ll := newLinkedList[int, int]()
	ll.addToFront(2, 102)
	ll.addToFront(3, 103)
	v, err := ll.popBack()
	assert.Nil(t, err)
	assert.Equal(t, v.key, 2)
	assert.Equal(t, v.value, 102)
}

func TestMakeEmptyAndAdd(t *testing.T) {
	ll := newLinkedList[int, int]()
	ll.addToFront(2, 102)
	ll.addToFront(3, 103)
	_, err := ll.popBack()
	assert.Nil(t, err)
	_, err = ll.popBack()
	assert.Nil(t, err)

	ll.addToBack(2, 102)
	ll.addToBack(3, 103)
	v, err := ll.popFront()
	assert.Nil(t, err)
	assert.Equal(t, v.key, 2)
	assert.Equal(t, v.value, 102)
}
