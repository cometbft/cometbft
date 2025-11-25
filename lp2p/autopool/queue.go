package autopool

import (
	"container/list"
	"errors"
	"sync"
)

// Simple thread-safe FIFO queue based on container/list.
// We can replace the implementation with a more efficient one if needed (ring buffer, MPMC, etc.)
type Queue struct {
	list *list.List
	mu   sync.RWMutex
}

var ErrPriority = errors.New("invalid priority")

// New Queue constructor.
func NewQueue() *Queue {
	return &Queue{
		list: list.New(),
		mu:   sync.RWMutex{},
	}
}

func (q *Queue) Len() int {
	q.mu.RLock()
	defer q.mu.RUnlock()

	return q.list.Len()
}

func (q *Queue) Push(value any) {
	q.mu.Lock()
	defer q.mu.Unlock()

	q.list.PushBack(value)
}

func (q *Queue) Pop() (any, bool) {
	q.mu.Lock()
	defer q.mu.Unlock()

	f := q.list.Front()
	if f == nil {
		return nil, false
	}

	value := f.Value

	q.list.Remove(f)

	return value, true
}

// PriorityQueue is a thread-safe queue of queues with a limited small number of priorities.
// Priority is an integer between 1 and the number of priorities (inclusive).
// Higher priority values are dequeued first.
type PriorityQueue struct {
	priorities           int
	levels               []*Queue
	highestNonEmptyLevel int
	mu                   sync.Mutex
}

func NewPriorityQueue(priorities int) *PriorityQueue {
	if priorities <= 0 {
		priorities = 1
	}

	queues := make([]*Queue, 0, priorities)
	for i := 0; i < priorities; i++ {
		queues = append(queues, NewQueue())
	}

	return &PriorityQueue{
		priorities:           priorities,
		levels:               queues,
		highestNonEmptyLevel: -1,
		mu:                   sync.Mutex{},
	}
}

func (q *PriorityQueue) Push(value any, priority int) error {
	if priority < 1 || priority > q.priorities {
		return ErrPriority
	}

	q.mu.Lock()
	defer q.mu.Unlock()

	idx := priority - 1

	q.levels[idx].Push(value)

	if idx > q.highestNonEmptyLevel {
		q.highestNonEmptyLevel = idx
	}

	return nil
}

func (q *PriorityQueue) Pop() (any, bool) {
	q.mu.Lock()
	defer q.mu.Unlock()

	// highest priority first
	for i := q.highestNonEmptyLevel; i >= 0; i-- {
		if v, ok := q.levels[i].Pop(); ok {
			q.updateHighestNonEmpty(i)
			return v, ok
		}
	}

	return nil, false
}

// updateHighestNonEmpty for empty PriorityQueue it will set highestNonEmptyLevel to -1
func (q *PriorityQueue) updateHighestNonEmpty(lastLevel int) {
	// noop
	if q.levels[lastLevel].Len() > 0 {
		return
	}

	// Update highestNonEmpty by scanning downward
	q.highestNonEmptyLevel = lastLevel - 1
	for q.highestNonEmptyLevel >= 0 && q.levels[q.highestNonEmptyLevel].Len() == 0 {
		q.highestNonEmptyLevel--
	}
}
