package autopool

import (
	"time"

	"github.com/Workiva/go-datastructures/queue"
)

const (
	// RBSize is an arbitrary default I just thought up.
	RBSize = 2048
	// GetTimeout is an arbitrary, but low, number I think we should timeout gets on
	GetTimeout = time.Millisecond * 5
)

// RBPriorityQueue is a thread-safe container of ring buffers with a limited small number of priorities.
// Priority is an integer between 1 and the number of priorities (inclusive).
// Higher priority values are dequeued first.
type RBPriorityQueue struct {
	priorities int
	levels     []*queue.RingBuffer
}

func NewRBPriorityQueue(priorities int) *RBPriorityQueue {
	if priorities <= 0 {
		priorities = 1
	}

	queues := make([]*queue.RingBuffer, 0, priorities)
	for i := 0; i < priorities; i++ {
		queues = append(queues, queue.NewRingBuffer(RBSize))
	}

	return &RBPriorityQueue{
		priorities: priorities,
		levels:     queues,
	}
}

func (q *RBPriorityQueue) Push(value any, priority int) error {
	if priority < 1 || priority > q.priorities {
		return ErrPriority
	}

	idx := priority - 1

	return q.levels[idx].Put(value)
}

func (q *RBPriorityQueue) Pop() (any, bool) {
	// highest priority first
	for i := q.priorities - 1; i >= 0; i-- {
		if q.levels[i].Len() > 0 {
			res, err := q.levels[i].Poll(GetTimeout)
			if err != nil {
				return nil, false
			}
			return res, true
		}
	}

	return nil, false
}
