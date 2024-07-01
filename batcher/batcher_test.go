package batcher

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestBatch(t *testing.T) {
	ch := make(chan int, 10)

	var count atomic.Int64
	fn := func(batch []int) {
		if len(batch) != 5 {
			t.Log("batch size not equal 5")
		}
		count.Add(int64(len(batch)))
	}

	ctx, cancel := context.WithCancel(context.Background())
	batch := New[int](ctx, 5, time.Second, fn, ch)
	go batch.RunLoop()

	for i := 0; i < 10; i++ {
		ch <- i
	}

	time.Sleep(time.Second)
	assert.Equal(t, int64(10), count.Load())

	for i := 0; i < 2; i++ {
		ch <- i
	}

	assert.Equal(t, int64(10), count.Load())
	time.Sleep(2 * time.Second)
	assert.Equal(t, int64(12), count.Load())
	cancel()
}
