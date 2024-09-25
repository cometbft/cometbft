package mempool

import (
	"context"
	"fmt"

	"github.com/cometbft/cometbft/internal/clist"
)

// IWRRIterator is the base struct for implementing iterators that traverse lanes with
// the Interleaved Weighted Round Robin (WRR) algorithm.
// https://en.wikipedia.org/wiki/Weighted_round_robin
type IWRRIterator struct {
	sortedLanes []lane
	laneIndex   int                        // current lane being iterated; index on sortedLanes
	cursors     map[LaneID]*clist.CElement // last accessed entries on each lane
	round       int                        // counts the rounds for IWRR
}

// This function picks the next lane to fetch an item from.
// If it was the last lane, it advances the round counter as well.
func (iter *IWRRIterator) advanceIndexes() lane {
	if iter.laneIndex == len(iter.sortedLanes)-1 {
		iter.round = (iter.round + 1) % (int(iter.sortedLanes[0].priority) + 1)
		if iter.round == 0 {
			iter.round++
		}
	}
	iter.laneIndex = (iter.laneIndex + 1) % len(iter.sortedLanes)
	return iter.sortedLanes[iter.laneIndex]
}

// Non-blocking version of the IWRR iterator to be used for reaping and
// rechecking transactions.
//
// This iterator does not support changes on the underlying mempool once initialized (or `Reset`),
// therefore the lock must be held on the mempool when iterating.
type NonBlockingIterator struct {
	IWRRIterator
}

func NewNonBlockingIterator(mem *CListMempool) *NonBlockingIterator {
	baseIter := IWRRIterator{
		sortedLanes: mem.sortedLanes,
		cursors:     make(map[LaneID]*clist.CElement, len(mem.lanes)),
		round:       1,
	}
	iter := &NonBlockingIterator{
		IWRRIterator: baseIter,
	}
	iter.reset(mem.lanes)
	return iter
}

// Reset must be called before every use of the iterator.
func (iter *NonBlockingIterator) reset(lanes map[LaneID]*clist.CList) {
	iter.laneIndex = 0
	iter.round = 1
	// Set cursors at the beginning of each lane.
	for lane := range lanes {
		iter.cursors[lane] = lanes[lane].Front()
	}
}

// Next returns the next element according to the WRR algorithm.
func (iter *NonBlockingIterator) Next() Entry {
	numEmptyLanes := 0

	lane := iter.sortedLanes[iter.laneIndex]
	for {
		// Skip empty lane or if cursor is at end of lane.
		if iter.cursors[lane.id] == nil {
			numEmptyLanes++
			if numEmptyLanes >= len(iter.sortedLanes) {
				return nil
			}
			lane = iter.advanceIndexes()
			continue
		}
		// Skip over-consumed lane on current round.
		if int(lane.priority) < iter.round {
			numEmptyLanes = 0
			lane = iter.advanceIndexes()
			continue
		}
		break
	}
	elem := iter.cursors[lane.id]
	if elem == nil {
		panic(fmt.Errorf("Iterator picked a nil entry on lane %s", lane.id))
	}
	iter.cursors[lane.id] = iter.cursors[lane.id].Next()
	_ = iter.advanceIndexes()
	return elem.Value.(*mempoolTx)
}

// BlockingIterator implements a blocking version of the WRR iterator,
// meaning that when no transaction is available, it will wait until a new one
// is added to the mempool.
// Unlike `NonBlockingIterator`, this iterator is expected to work with an evolving mempool.
type BlockingIterator struct {
	IWRRIterator
	ctx  context.Context
	mp   *CListMempool
	name string // for debugging
}

func NewBlockingIterator(ctx context.Context, mem *CListMempool, name string) Iterator {
	iter := IWRRIterator{
		sortedLanes: mem.sortedLanes,
		cursors:     make(map[LaneID]*clist.CElement, len(mem.sortedLanes)),
		round:       1,
	}
	return &BlockingIterator{
		IWRRIterator: iter,
		ctx:          ctx,
		mp:           mem,
		name:         name,
	}
}

// WaitNextCh returns a channel to wait for the next available entry. The channel will be explicitly
// closed when the entry gets removed before it is added to the channel, or when reaching the end of
// the list.
//
// Unsafe for concurrent use by multiple goroutines.
func (iter *BlockingIterator) WaitNextCh() <-chan Entry {
	ch := make(chan Entry)
	go func() {
		var lane lane
		for {
			l, addTxCh := iter.pickLane()
			if addTxCh == nil {
				lane = l
				break
			}
			// There are no transactions to take from any lane. Wait until at
			// least one is added to the mempool and try again.
			select {
			case <-addTxCh:
			case <-iter.ctx.Done():
				close(ch)
				return
			}
		}
		if elem := iter.next(lane.id); elem != nil {
			ch <- elem.Value.(Entry)
		}
		// Unblock receiver in case no entry was sent (it will receive nil).
		close(ch)
	}()
	return ch
}

// pickLane returns a _valid_ lane on which to iterate, according to the WRR
// algorithm. A lane is valid if it is not empty and it is not over-consumed,
// meaning that the number of accessed entries in the lane has not yet reached
// its priority value in the current WRR iteration. It returns a channel to wait
// for new transactions if all lanes are empty or don't have transactions that
// have not yet been accessed.
func (iter *BlockingIterator) pickLane() (lane, chan struct{}) {
	iter.mp.addTxChMtx.RLock()
	defer iter.mp.addTxChMtx.RUnlock()

	// Start from the last accessed lane.
	currLane := iter.sortedLanes[iter.laneIndex]

	// Loop until finding a valid lane. If the current lane is not valid,
	// continue with the next lower-priority lane, in a round robin fashion.
	numEmptyLanes := 0
	for {
		laneID := currLane.id
		// Skip empty lanes or lanes with their cursor pointing at their last entry.
		if iter.mp.lanes[laneID].Len() == 0 ||
			(iter.cursors[laneID] != nil &&
				iter.cursors[laneID].Value.(*mempoolTx).seq == iter.mp.addTxLaneSeqs[laneID]) {
			numEmptyLanes++
			if numEmptyLanes >= len(iter.sortedLanes) {
				// There are no lanes with non-accessed entries. Wait until a
				// new tx is added.
				return lane{}, iter.mp.addTxCh
			}
			currLane = iter.advanceIndexes()
			continue
		}

		// Skip over-consumed lanes.
		if int(currLane.priority) < iter.round {
			numEmptyLanes = 0
			currLane = iter.advanceIndexes()
			continue
		}

		_ = iter.advanceIndexes()
		return currLane, nil
	}
}

// In classical WRR, the iterator cycles over the lanes. When a lane is selected, Next returns an
// entry from the selected lane. On subsequent calls, Next will return the next entries from the
// same lane until `lane` entries are accessed or the lane is empty, where `lane` is the priority.
// The next time, Next will select the successive lane with lower priority.
// next returns the next entry from the given lane and updates WRR variables.
func (iter *BlockingIterator) next(laneID LaneID) *clist.CElement {
	// Load the last accessed entry in the lane and set the next one.
	var next *clist.CElement

	if cursor := iter.cursors[laneID]; cursor != nil {
		// If the current entry is the last one or was removed, Next will return nil.
		// Note we don't need to wait until the next entry is available (with <-cursor.NextWaitChan()).
		next = cursor.Next()
	} else {
		// We are at the beginning of the iteration or the saved entry got removed. Pick the first
		// entry in the lane if it's available (don't wait for it); if not, Front will return nil.
		next = iter.mp.lanes[laneID].Front()
	}

	// Update auxiliary variables.
	if next != nil {
		// Save entry.
		iter.cursors[laneID] = next
	} else {
		// The entry got removed or it was the last one in the lane.
		// At the moment this should not happen - the loop in PickLane will loop forever until there
		// is data in at least one lane
		delete(iter.cursors, laneID)
	}

	return next
}
