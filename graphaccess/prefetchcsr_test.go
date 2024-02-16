package graphaccess

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

// Channel creation takes about 40 ns.
// Future creation takes about 29 ns.
func BenchmarkFutureCreation(b *testing.B) {
	var _ future[int]
	for i := 0; i < b.N; i++ {
		_ = newFuture[int]()
	}
}

func TestMarshalling(t *testing.T) {
	stats := PrefetchStats{}
	assert.Equal(t, "{\"S3Fetches\":0,\"cacheHits\":0,\"inFlightHits\":0,\"prefetcherHits\":0}",
		stats.convertToString())
}
