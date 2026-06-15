package blocksync

import (
	"time"

	"github.com/cometbft/cometbft/internal/movavg"
)

// SyncTracker determines when to escape blocksync and switch to consensus,
// using only information that cannot be forged by a malicious peer:
//   - wall-clock time (local timer)
//   - block header timestamps (from cryptographically validated blocks)
//
// SyncTracker does NOT have its own mutex — it is always used within the
// BlockPool's mutex scope.
//
// Strategy:
//
//	After at least one block has been synced, we compare the wall-clock
//	receive interval against the block production interval (moving average
//	of recent header timestamp gaps). If receiving a block takes as long
//	or longer than the network took to produce it, we're not catching up
//	and should switch to consensus.
//
//	A fixed minimum timeout (noBlockTimeout) serves as the floor: regardless
//	of production speed, if no block arrives within noBlockTimeout, we
//	declare ourselves caught up. This catches peer stall and very fast
//	production chains.
type SyncTracker struct {
	noBlockTimeout time.Duration

	lastBlockWallTime   time.Time
	lastBlockHeaderTime time.Time
	prodMA              *movavg.MovingAverage // moving average of production intervals
}

// NewSyncTracker creates a SyncTracker. timeout is the maximum time to wait
// without a block before declaring caught up (the floor, always used).
func NewSyncTracker(timeout time.Duration) *SyncTracker {
	return &SyncTracker{
		noBlockTimeout: timeout,
		prodMA:         movavg.NewMovingAverage(10), // window of 10 production intervals
	}
}

// RecordBlock records a successfully synced block. Caller must hold BlockPool mutex.
func (s *SyncTracker) RecordBlock(blockHeaderTime time.Time) {
	now := time.Now()
	if !s.lastBlockHeaderTime.IsZero() {
		interval := blockHeaderTime.Sub(s.lastBlockHeaderTime)
		if interval > 0 {
			s.prodMA.Add(interval)
		}
	}
	s.lastBlockHeaderTime = blockHeaderTime
	s.lastBlockWallTime = now
}

// recordBlockAt is like RecordBlock but uses an explicit wall clock.
// Intended for tests that simulate time.
func (s *SyncTracker) recordBlockAt(blockHeaderTime time.Time, wallTime time.Time) {
	if !s.lastBlockHeaderTime.IsZero() {
		interval := blockHeaderTime.Sub(s.lastBlockHeaderTime)
		if interval > 0 {
			s.prodMA.Add(interval)
		}
	}
	s.lastBlockHeaderTime = blockHeaderTime
	s.lastBlockWallTime = wallTime
}

// IsCaughtUp checks whether to escape blocksync. Caller must hold BlockPool mutex.
func (s *SyncTracker) IsCaughtUp() bool {
	return s.isCaughtUpAt(time.Now())
}

// IsCaughtUpAt is like IsCaughtUp but uses the given time as "now".
func (s *SyncTracker) IsCaughtUpAt(now time.Time) bool {
	return s.isCaughtUpAt(now)
}

func (s *SyncTracker) isCaughtUpAt(now time.Time) bool {
	receiveInterval := now.Sub(s.lastBlockWallTime)

	prodInterval, hasProd := s.prodMA.Avg()

	// Regime 1: Far behind (block age >> production interval).
	// Tolerate larger gaps proportional to how far behind we are.
	// Requires >= 3 production samples, same as Regime 2, to avoid a single
	// outlier block (e.g., chain restart after a long pause) from producing
	// a wildly inflated timeout.
	if hasProd && s.prodMA.Len() >= 3 {
		blockAge := now.Sub(s.lastBlockHeaderTime)
		if blockAge > prodInterval*100 {
			timeout := blockAge / 100
			if timeout < s.noBlockTimeout {
				timeout = s.noBlockTimeout
			}
			return receiveInterval >= timeout
		}
	}

	// Regime 2: Near tip. Smooth transition from aggressive rate check
	// to the noBlockTimeout floor as we get further from the tip.
	//   At blockAge = 0:              threshold = prodInterval × 0.6
	//   At blockAge = prodInterval × 100: threshold = noBlockTimeout
	// This avoids a hard cliff at the regime boundary.
	if hasProd && s.prodMA.Len() >= 3 {
		ratio := float64(now.Sub(s.lastBlockHeaderTime)) / float64(prodInterval*100)
		ratio = max(0.0, min(1.0, ratio))
		// threshold = prodInterval*0.6 + ratio*(noBlockTimeout - prodInterval*0.6)
		threshold := time.Duration(float64(prodInterval)*0.6 +
			ratio*(float64(s.noBlockTimeout)-float64(prodInterval)*0.6))
		return receiveInterval >= threshold
	}

	return receiveInterval >= s.noBlockTimeout
}
