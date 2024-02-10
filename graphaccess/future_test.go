package graphaccess

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestFuture(t *testing.T) {
	f := newFuture[int]()
	go func() {
		f.put(22)
	}()
	assert.Equal(t, 22, f.get())
	assert.Equal(t, 22, f.get())
}
