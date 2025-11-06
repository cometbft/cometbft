package priorityqueue

import (
	"sync"
	"time"
)

// Queue is a priority queue implementation that supports
// push and pop operations. The queue is thread-safe and can be used
type Queue struct {
	container *container
	mu        sync.RWMutex
}

// New Queue constructor
func New(cap int) *Queue {
	return &Queue{
		container: newContainer(true, cap),
		mu:        sync.RWMutex{},
	}
}

func (q *Queue) Push(v any, priority uint64) {
	q.mu.Lock()
	defer q.mu.Unlock()

	q.container.PushItem(v, priority)
}

func (q *Queue) Pop() (v any, ok bool) {
	q.mu.Lock()
	defer q.mu.Unlock()

	return q.container.PopItem()
}

func (q *Queue) Consumer() (<-chan any, func()) {
	var (
		out        = make(chan any)
		cancelChan = make(chan struct{})
		cancel     = func() { close(cancelChan) }
	)

	go func() {
		for {
			select {
			case <-cancelChan:
				close(out)
				return
			default:
				// non blocking, just continue
			}

			v, ok := q.Pop()
			if !ok {
				time.Sleep(10 * time.Millisecond)
				continue
			}

			out <- v
		}

	}()

	return out, cancel
}
