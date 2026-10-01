package pool_test

import (
	"alla/worker-service/internal/pool"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestPoolLimitsConcurrency(t *testing.T) {
	const size = 3
	const tasks = 20

	p := pool.New(size)

	var current, max atomic.Int64

	for range tasks {
		p.Submit(func() {
			n := current.Add(1)
			for {
				m := max.Load()
				if n <= m || max.CompareAndSwap(m, n) {
					break
				}
			}
			time.Sleep(20 * time.Millisecond)
			current.Add(-1)
		})
	}

	p.Stop()

	assert.Equal(t, int64(size), max.Load())
	assert.Equal(t, int64(0), current.Load())
}

func TestPoolStopWaitsForTasks(t *testing.T) {
	const size = 2
	const tasks = 10

	p := pool.New(size)

	var done atomic.Int64

	for range tasks {
		p.Submit(func() {
			time.Sleep(5 * time.Millisecond)
			done.Add(1)
		})
	}

	p.Stop()

	assert.Equal(t, int64(tasks), done.Load())
}
