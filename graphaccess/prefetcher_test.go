package graphaccess

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func fetcher(num uint32) []edge {
	time.Sleep(10 * time.Millisecond)
	return []edge{{1, num}, {1, num + 1}, {1, num + 3}}
}

func TestPrefetchFunctionality(t *testing.T) {
	pf := NewPrefetcher(2, 10, fetcher)
	go func() {
		for i := 1; i <= 10; i++ {
			pf.write([]uint32{uint32(i)})
		}
	}()
	for i := 1; i <= 10; i += 2 {
		//i and i+1 should be in flight
		f, found := pf.getFromInFlightQueue(uint32(i))
		for !found {
			f, found = pf.getFromInFlightQueue(uint32(i))
		}
		f1, found := pf.getFromInFlightQueue(uint32(i + 1))
		for !found {
			f1, found = pf.getFromInFlightQueue(uint32(i + 1))
		}
		f.get()
		f1.get()
	}
	for i := 1; i <= 10; i++ {
		v, found := pf.getFromPrefetchCache(uint32(i))
		assert.True(t, found)
		assert.Equal(t, 3, len(v))
	}
}
