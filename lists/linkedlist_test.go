package lists

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestLinkedListStack(t *testing.T) {
	ll := NewLinkedList[int, int]()
	ll.AddToFront(2, 102)
	ll.AddToFront(3, 103)
	v, err := ll.PopFront()
	assert.Nil(t, err)
	assert.Equal(t, v.Key, 3)
	assert.Equal(t, v.Value, 103)
}

func TestLinkedListQueue(t *testing.T) {
	ll := NewLinkedList[int, int]()
	ll.AddToFront(2, 102)
	ll.AddToFront(3, 103)
	v, err := ll.PopBack()
	assert.Nil(t, err)
	assert.Equal(t, v.Key, 2)
	assert.Equal(t, v.Value, 102)
}

func TestMakeEmptyAndAdd(t *testing.T) {
	ll := NewLinkedList[int, int]()
	ll.AddToFront(2, 102)
	ll.AddToFront(3, 103)
	_, err := ll.PopBack()
	assert.Nil(t, err)
	_, err = ll.PopBack()
	assert.Nil(t, err)

	ll.AddToBack(2, 102)
	ll.AddToBack(3, 103)
	v, err := ll.PopFront()
	assert.Nil(t, err)
	assert.Equal(t, v.Key, 2)
	assert.Equal(t, v.Value, 102)
}
