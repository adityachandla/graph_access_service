package graphaccess

import "testing"

// Channel creation takes about 40 ns.
// Future creation takes about 29 ns.
func BenchmarkFutureCreation(b *testing.B) {
	var _ future[int]
	for i := 0; i < b.N; i++ {
		_ = newFuture[int]()
	}
}
