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
	size    int
	mu      sync.Mutex

	// valuesAvailable signals that there are values available in one of the
	// queues that can be popped
	valuesAvailable chan struct{}
}

func NewPriorityQueue(priorities int) *PriorityQueue {
	return NewPriorityQueueWithMax(priorities, 0)
}

// NewPriorityQueueWithMax bounds total depth at maxSize (0 means unbounded).
// Once full, Push evicts the oldest strictly-lower-priority item to admit a
// higher-priority one; ErrQueueFull is only returned when no lower-priority
// item is available to evict.
func NewPriorityQueueWithMax(priorities, maxSize int) *PriorityQueue {
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
		maxSize:              maxSize,
		mu:                   sync.Mutex{},
		valuesAvailable:      make(chan struct{}, 1),
	}
}

func (q *PriorityQueue) Push(value any, priority int) error {
	if priority < 1 || priority > q.priorities {
		return ErrPriority
	}

	q.mu.Lock()
	defer q.mu.Unlock()

	idx := priority - 1

	// maxSize caps total depth across all levels. Once full, make room by
	// dropping the oldest strictly-lower-priority item instead of rejecting
	// this push outright, so a burst of low-priority traffic can't starve
	// higher-priority messages out of admission. If nothing lower is queued,
	// fall back to rejecting, same as before.
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

// evictLowerPriority drops the oldest queued item from the lowest occupied
// level below newIdx, making room for an admission at newIdx. Returns false
// if no strictly-lower-priority item is queued to evict.
func (q *PriorityQueue) evictLowerPriority(newIdx int) bool {
	for i := 0; i < newIdx; i++ {
		if _, ok := q.levels[i].Pop(); ok {
			q.size--
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
