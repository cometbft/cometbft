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

var (
	ErrPriority  = errors.New("invalid priority")
	ErrQueueFull = errors.New("priority queue is full")
)

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
	// maxSize is the total capacity across all priority levels; 0 means unlimited.
	maxSize int
	// size mirrors sum(levels[i].Len()); kept as a counter so Push doesn't need
	// to scan all levels (and their locks) to check capacity.
	size int
	mu   sync.Mutex

	// onEvict, if set, is called with the value dropped by evictLowerPriority.
	onEvict func(value any)

	// valuesAvailable signals that there are values available in one of the
	// queues that can be popped
	valuesAvailable chan struct{}
}

// PriorityQueueOption configures a PriorityQueue at construction time.
type PriorityQueueOption func(*PriorityQueue)

// WithOnEvict sets a callback invoked with the value dropped whenever Push
// evicts a lower-priority item to make room.
func WithOnEvict(onEvict func(value any)) PriorityQueueOption {
	return func(q *PriorityQueue) { q.onEvict = onEvict }
}

func NewPriorityQueue(priorities int) *PriorityQueue {
	return NewPriorityQueueWithMax(priorities, 0)
}

// NewPriorityQueueWithMax bounds total depth at maxSize (0 means unbounded).
// Once full, Push evicts a lower-priority item rather than reject, so low-priority bursts can't starve high-priority admission.
func NewPriorityQueueWithMax(priorities, maxSize int, opts ...PriorityQueueOption) *PriorityQueue {
	if priorities <= 0 {
		priorities = 1
	}

	queues := make([]*Queue, 0, priorities)
	for i := 0; i < priorities; i++ {
		queues = append(queues, NewQueue())
	}

	q := &PriorityQueue{
		priorities:           priorities,
		levels:               queues,
		highestNonEmptyLevel: -1,
		maxSize:              maxSize,
		mu:                   sync.Mutex{},
		valuesAvailable:      make(chan struct{}, 1),
	}

	for _, opt := range opts {
		opt(q)
	}

	return q
}

func (q *PriorityQueue) Push(value any, priority int) error {
	if priority < 1 || priority > q.priorities {
		return ErrPriority
	}

	q.mu.Lock()
	defer q.mu.Unlock()

	idx := priority - 1

	if q.maxSize > 0 && q.size >= q.maxSize && !q.evictLowerPriority(idx) {
		return ErrQueueFull
	}

	q.levels[idx].Push(value)
	q.size++

	if idx > q.highestNonEmptyLevel {
		q.highestNonEmptyLevel = idx
	}

	q.notifyValuesAvailable()

	return nil
}

// evictLowerPriority drops the oldest item from the lowest occupied level below newIdx; false if none to evict.
func (q *PriorityQueue) evictLowerPriority(newIdx int) bool {
	for i := 0; i < newIdx; i++ {
		if v, ok := q.levels[i].Pop(); ok {
			q.size--
			if q.onEvict != nil {
				q.onEvict(v)
			}
			return true
		}
	}
	return false
}

// notifyValuesAvailable notifies callers waiting on the channel returned by
// WaitForValues that work is values are available to be popped.
func (q *PriorityQueue) notifyValuesAvailable() {
	// if there is already a value in the channel (valuesAvailable is buffered
	// with cap 1), then do nothing, since there is already a notification
	// waiting to be pulled off the channel telling callers that values are
	// available
	select {
	case q.valuesAvailable <- struct{}{}:
	default:
	}
}

func (q *PriorityQueue) Pop() (any, bool) {
	q.mu.Lock()
	defer q.mu.Unlock()

	// highest priority first
	for i := q.highestNonEmptyLevel; i >= 0; i-- {
		if v, ok := q.levels[i].Pop(); ok {
			q.size--
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

// WaitForValues returns a channel that will have a value in it when new values
// are ready to be popped from the priority queue.
func (q *PriorityQueue) WaitForValues() <-chan struct{} {
	return q.valuesAvailable
}
