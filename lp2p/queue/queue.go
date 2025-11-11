package queue

import (
	"container/list"
	"errors"
	"sync"
)

// Simple thread-safe FIFO queue based on container/list.
// We can replace the implementation with a more efficient one if needed (ring buffer, MPMC, etc.)
type Queue struct {
	list *list.List
	mu   sync.Mutex
}

var ErrPriority = errors.New("invalid priority")

// New Queue constructor.
func New() *Queue {
	return &Queue{
		list: list.New(),
		mu:   sync.Mutex{},
	}
}

func (q *Queue) Len() int {
	q.mu.Lock()
	defer q.mu.Unlock()

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
	priorities int
	levels     []*Queue
	mu         sync.Mutex
}

func NewPriorityQueue(priorities int) *PriorityQueue {
	if priorities <= 0 {
		priorities = 1
	}

	queues := make([]*Queue, 0, priorities)
	for i := 0; i < priorities; i++ {
		queues = append(queues, New())
	}

	return &PriorityQueue{
		priorities: priorities,
		levels:     queues,
		mu:         sync.Mutex{},
	}
}

func (q *PriorityQueue) Push(value any, priority int) error {
	if priority < 1 || priority > q.priorities {
		return ErrPriority
	}

	q.levels[priority-1].Push(value)

	return nil
}

func (q *PriorityQueue) Pop() (any, bool) {
	// edge case: only one priority
	if q.priorities == 1 {
		return q.levels[0].Pop()
	}

	q.mu.Lock()
	defer q.mu.Unlock()

	// highest priority first
	for i := q.priorities - 1; i >= 0; i-- {
		if v, ok := q.levels[i].Pop(); ok {
			return v, ok
		}
	}

	return nil, false
}
