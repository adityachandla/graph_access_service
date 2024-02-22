package lists_test

import (
	"fmt"
	"github.com/adityachandla/graph_access_service/lists"
	"github.com/stretchr/testify/assert"
	"sync"
	"testing"
)

func TestNormalAddition(t *testing.T) {
	cq := lists.NewDFSQueue[int](4)
	cq.WriteAll([]int{2, 3})

	v := cq.Read()
	assert.Equal(t, 2, v)

	v = cq.Read()
	assert.Equal(t, 3, v)
}

func TestOverwrite(t *testing.T) {
	cq := lists.NewDFSQueue[int](4)
	cq.WriteAll([]int{4, 3, 2, 1})
	cq.WriteAll([]int{5, 6})
	readRes := []int{5, 6, 4, 3}
	for _, v := range readRes {
		value := cq.Read()
		assert.Equal(t, v, value)
	}
}

func TestOverwrite2(t *testing.T) {
	cq := lists.NewDFSQueue[int](4)
	cq.WriteAll([]int{4, 3, 2, 1})
	cq.WriteAll([]int{5, 6})
	cq.WriteAll([]int{1})
	readRes := []int{1, 5, 6, 4}
	for _, v := range readRes {
		value := cq.Read()
		assert.Equal(t, v, value)
	}
}

func TestMultithreading(t *testing.T) {
	cq := lists.NewDFSQueue[int](5)
	wg := sync.WaitGroup{}
	wg.Add(5)
	for i := 0; i < 5; i++ {
		go func(v int) {
			defer wg.Done()
			val := cq.Read()
			fmt.Printf("%d read %d\n", v, val)
		}(i)
	}
	cq.WriteAll([]int{5, 4, 3, 2, 1})
	wg.Wait()
}

func TestBfsQueue(t *testing.T) {
	bq := lists.NewBFSQueue[int](4)
	bq.WriteAll([]int{1, 2})
	bq.Read()
}
