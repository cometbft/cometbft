package mempool

import (
	"fmt"

	"github.com/cometbft/cometbft/internal/clist"
	"github.com/cometbft/cometbft/types"
)

// WRRIterator is the base struct for implementing iterators that traverse lanes with
// the classical Weighted Round Robin (WRR) algorithm.
type WRRIterator struct {
	sortedLanes []types.Lane
	laneIndex   int                            // current lane being iterated; index on sortedLanes
	counters    map[types.Lane]uint            // counters of consumed entries, for WRR algorithm
	cursors     map[types.Lane]*clist.CElement // last accessed entries on each lane
}

func (iter *WRRIterator) nextLane() types.Lane {
	iter.laneIndex = (iter.laneIndex + 1) % len(iter.sortedLanes)
	return iter.sortedLanes[iter.laneIndex]
}

// Non-blocking version of the WRR iterator to be used for reaping and
// rechecking transactions.
//
// Lock must be held on the mempool when iterating: the mempool cannot be
// modified while iterating.
type NonBlockingWRRIterator struct {
	WRRIterator
}

func NewWRRIterator(mem *CListMempool) *NonBlockingWRRIterator {
	baseIter := WRRIterator{
		sortedLanes: mem.sortedLanes,
		counters:    make(map[types.Lane]uint, len(mem.lanes)),
		cursors:     make(map[types.Lane]*clist.CElement, len(mem.lanes)),
	}
	iter := &NonBlockingWRRIterator{
		WRRIterator: baseIter,
	}
	iter.Reset(mem.lanes)
	return iter
}

// Reset must be called before every use of the iterator.
func (iter *NonBlockingWRRIterator) Reset(lanes map[types.Lane]*clist.CList) {
	iter.laneIndex = 0
	for i := range iter.counters {
		iter.counters[i] = 0
	}
	// Set cursors at the beginning of each lane.
	for lane := range lanes {
		iter.cursors[lane] = lanes[lane].Front()
	}
}

// Next returns the next element according to the WRR algorithm.
func (iter *NonBlockingWRRIterator) Next() Entry {
	lane := iter.sortedLanes[iter.laneIndex]
	numEmptyLanes := 0
	for {
		// Skip empty lane or if cursor is at end of lane.
		if iter.cursors[lane] == nil {
			numEmptyLanes++
			if numEmptyLanes >= len(iter.sortedLanes) {
				return nil
			}
			lane = iter.nextLane()
			continue
		}
		// Skip over-consumed lane.
		if iter.counters[lane] >= uint(lane) {
			iter.counters[lane] = 0
			numEmptyLanes = 0
			lane = iter.nextLane()
			continue
		}
		break
	}
	elem := iter.cursors[lane]
	if elem == nil {
		panic(fmt.Errorf("Iterator picked a nil entry on lane %d", lane))
	}
	iter.cursors[lane] = iter.cursors[lane].Next()
	iter.counters[lane]++
	return elem.Value.(*mempoolTx)
}

// BlockingWRRIterator implements a blocking version of the WRR iterator,
// meaning that when no transaction is available, it will wait until a new one
// is added to the mempool.
type BlockingWRRIterator struct {
	WRRIterator
	mp *CListMempool
}

func NewBlockingWRRIterator(mem *CListMempool) Iterator {
	iter := WRRIterator{
		sortedLanes: mem.sortedLanes,
		counters:    make(map[types.Lane]uint, len(mem.sortedLanes)),
		cursors:     make(map[types.Lane]*clist.CElement, len(mem.sortedLanes)),
	}
	return &BlockingWRRIterator{
		WRRIterator: iter,
		mp:          mem,
	}
}

// WaitNextCh returns a channel to wait for the next available entry. The channel will be explicitly
// closed when the entry gets removed before it is added to the channel, or when reaching the end of
// the list.
//
// Unsafe for concurrent use by multiple goroutines.
func (iter *BlockingWRRIterator) WaitNextCh() <-chan Entry {
	ch := make(chan Entry)
	go func() {
		// Add the next entry to the channel if not nil.
		lane := iter.PickLane()
		if entry := iter.Next(lane); entry != nil {
			ch <- entry.Value.(Entry)
			close(ch)
		} else {
			// Unblock the receiver (it will receive nil).
			close(ch)
		}
	}()
	return ch
}

// PickLane returns a _valid_ lane on which to iterate, according to the WRR
// algorithm. A lane is valid if it is not empty or it is not over-consumed,
// meaning that the number of accessed entries in the lane has not yet reached
// its priority value in the current WRR iteration. It will block until a
// transaction is available in any lane.
func (iter *BlockingWRRIterator) PickLane() types.Lane {
	// Start from the last accessed lane.
	lane := iter.sortedLanes[iter.laneIndex]

	iter.mp.addTxChMtx.RLock()
	defer iter.mp.addTxChMtx.RUnlock()

	// Loop until finding a valid lane. If the current lane is not valid,
	// continue with the next lower-priority lane, in a round robin fashion.
	numEmptyLanes := 0
	for {
		// Skip empty lanes or lanes with their cursor pointing at their last entry.
		if iter.mp.lanes[lane].Len() == 0 ||
			(iter.cursors[lane] != nil &&
				iter.cursors[lane].Value.(*mempoolTx).seq == iter.mp.addTxLaneSeqs[lane]) {
			numEmptyLanes++
			if numEmptyLanes >= len(iter.sortedLanes) {
				// There are no lanes with non-accessed entries. Wait until a new tx is added.
				ch := iter.mp.addTxCh
				iter.mp.addTxChMtx.RUnlock()
				<-ch
				iter.mp.addTxChMtx.RLock()
				numEmptyLanes = 0
			}
			lane = iter.nextLane()
			continue
		}

		// Skip over-consumed lanes.
		if iter.counters[lane] >= uint(lane) {
			iter.counters[lane] = 0
			numEmptyLanes = 0
			lane = iter.nextLane()
			continue
		}

		return lane
	}
}

// Next returns the next element according to the WRR algorithm.
//
// In classical WRR, the iterator cycles over the lanes. When a lane is selected, Next returns an
// entry from the selected lane. On subsequent calls, Next will return the next entries from the
// same lane until `lane` entries are accessed or the lane is empty, where `lane` is the priority.
// The next time, Next will select the successive lane with lower priority.
func (iter *BlockingWRRIterator) Next(lane types.Lane) *clist.CElement {
	// Load the last accessed entry in the lane and set the next one.
	var next *clist.CElement
	if cursor := iter.cursors[lane]; cursor != nil {
		// If the current entry is the last one or was removed, Next will return nil.
		// Note we don't need to wait until the next entry is available (with <-cursor.NextWaitChan()).
		next = cursor.Next()
	} else {
		// We are at the beginning of the iteration or the saved entry got removed. Pick the first
		// entry in the lane if it's available (don't wait for it); if not, Front will return nil.
		next = iter.mp.lanes[lane].Front()
	}

	// Update auxiliary variables.
	if next != nil {
		// Save entry and increase the number of accessed transactions for this lane.
		iter.cursors[lane] = next
		iter.counters[lane]++
	} else {
		// The entry got removed or it was the last one in the lane.
		// At the moment this should not happen - the loop in PickLane will loop forever until there
		// is data in at least one lane
		delete(iter.cursors, lane)
	}

	return next
}