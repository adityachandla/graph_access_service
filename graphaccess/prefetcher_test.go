package graphaccess

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func fetcher(num uint32) edgeList {
	time.Sleep(10 * time.Millisecond)
	return []byte{1, 0, 0, 0, byte(num), 0, 0, 0,
		1, 0, 0, 0, byte(num + 1), 0, 0, 0,
		1, 0, 0, 0, byte(num + 3), 0, 0, 0}
}

func TestPrefetchFunctionality(t *testing.T) {
	pf := NewPrefetcher(BFS, fetcher)
	go func() {
		for i := 1; i <= 10; i++ {
			pf.write([]uint32{uint32(i)})
		}
	}()
	for j := 0; j < 2; j++ {
		futures := make([]*future[edgeList], 5)
		for i := 1; i <= 5; i++ {
			f, found := pf.getFromInFlightQueue(uint32((5 * j) + i))
			futures[i-1] = f
			for !found {
				futures[i-1], found = pf.getFromInFlightQueue(uint32((5 * j) + i))
			}
		}
		for _, f := range futures {
			f.get()
		}
	}
	for i := 1; i <= 10; i++ {
		v, found := pf.getFromPrefetchCache(uint32(i))
		assert.True(t, found)
		assert.Equal(t, 3, v.Len())
	}
}
