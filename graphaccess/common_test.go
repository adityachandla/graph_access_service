package graphaccess

import (
	"testing"
)

func TestEdgesMiddle(t *testing.T) {
	edgeList := []edge{{1, 2}, {1, 4}, {2, 3}, {2, 4}, {2, 5}, {3, 1}, {3, 3}}
	res := getEdgesWithLabel(edgeList[0:], 2)
	if !arrayEqual(res, []uint32{3, 4, 5}) {
		t.Fail()
	}
	res = getEdgesWithLabel(edgeList[0:4], 2)
	if !arrayEqual(res, []uint32{3, 4}) {
		t.Fail()
	}
	res = getEdgesWithLabel(edgeList[3:4], 2)
	if !arrayEqual(res, []uint32{4}) {
		t.Fail()
	}
	res = getEdgesWithLabel(edgeList[0:3], 2)
	if !arrayEqual(res, []uint32{3}) {
		t.Fail()
	}
	res = getEdgesWithLabel(edgeList[4:], 2)
	if !arrayEqual(res, []uint32{5}) {
		t.Fail()
	}
}

func arrayEqual[T comparable](one, two []T) bool {
	if len(one) != len(two) {
		return false
	}
	for i := 0; i < len(one); i++ {
		if one[i] != two[i] {
			return false
		}
	}
	return true
}
