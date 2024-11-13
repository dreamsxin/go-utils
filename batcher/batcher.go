package batcher

import (
	"context"
	//"log"
	"time"
)

type Batcher[T any] struct {
	ctx       context.Context
	batchSize int
	wait      time.Duration
	fn        func([]T)
	ch        <-chan T
}

func New[T any](ctx context.Context, batchSize int, wait time.Duration, fn func([]T), ch <-chan T) Batcher[T] {
	if fn == nil {
		panic("fn is nil")
	}
	if ch == nil {
		panic("ch is nil")
	}
	return Batcher[T]{ctx, batchSize, wait, fn, ch}
}

func (t Batcher[T]) Close() {
	t.ctx.Done()
}

// Batch reads from a channel and calls fn with a slice of batchSize.
func (t Batcher[T]) RunLoop() {
	if t.batchSize <= 1 {
		for v := range t.ch {
			t.fn([]T{v})
		}

	} else {
		ticker := time.NewTicker(t.wait)
		defer ticker.Stop()
		var batch = make([]T, 0, t.batchSize)
		for {
			select {
			case <-t.ctx.Done():
				//log.Default().Println("close")
				if len(batch) > 0 {
					t.fn(batch)
				}
				return
			case v, ok := <-t.ch:
				//log.Default().Println("get")
				if !ok { // closed
					t.fn(batch)
					return
				}

				batch = append(batch, v)
				if len(batch) == t.batchSize { // full
					t.fn(batch)
					batch = make([]T, 0, t.batchSize) // reset
				}
			case <-ticker.C:
				//log.Default().Println("ticker")
				if len(batch) > 0 { // partial
					t.fn(batch)
					batch = make([]T, 0, t.batchSize) // reset
				}
			}
		}
	}
}
