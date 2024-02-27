package graphaccess

import (
	"github.com/adityachandla/graph_access_service/bin_util"
)

type edgeList []byte

func (e edgeList) LabelAt(index int) uint32 {
	//Label is the first one.
	actualIndex := index * 8
	return bin_util.ByteToUint(e[actualIndex : actualIndex+4])
}

func (e edgeList) NodeAt(index int) uint32 {
	//Label is the second one.
	actualIndex := (index * 8) + 4
	return bin_util.ByteToUint(e[actualIndex : actualIndex+4])
}

func (e edgeList) SliceStart(start int) edgeList {
	//start index is inclusive.
	return e[start*8:]
}

func (e edgeList) SliceEnd(end int) edgeList {
	//end index is not inclusive.
	return e[:end*8]
}

func (e edgeList) Len() int {
	return len(e) / 8
}
