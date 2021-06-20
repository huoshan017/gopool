package gopool

import (
	"context"
	"sync"
	"testing"
)

func TestGoPool(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	pool := NewPool(10)
	count := 10000

	wg := &sync.WaitGroup{}
	wg.Add(count)

	for i := 0; i < count; i++ {
		pool.CommitTask(ctx, func(a interface{}) interface{} {
			aa := a.(int)
			t.Logf("ok %v", aa)
			wg.Done()
			return nil
		}, i)
	}

	wg.Wait()
}
