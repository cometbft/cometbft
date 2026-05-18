package blocksync

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func fakeNow() time.Time {
	return time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
}

func at(offset time.Duration) time.Time {
	return fakeNow().Add(offset)
}

func TestSyncTrackerTimeoutFloor(t *testing.T) {
	// Before any production info is available (single block),
	// uses the noBlockTimeout floor.
	s := NewSyncTracker(1 * time.Second)

	s.recordBlockAt(at(0), at(0))

	// 500ms later: receive < timeout → not caught up.
	require.False(t, s.IsCaughtUpAt(at(500*time.Millisecond)))

	// 1s later: receive >= timeout → caught up.
	require.True(t, s.IsCaughtUpAt(at(1*time.Second)))
}

func TestSyncTrackerCatchingUpFasterThanProduction(t *testing.T) {
	// Receiving faster than production: not caught up.
	s := NewSyncTracker(5 * time.Second)

	s.recordBlockAt(at(0), at(0))
	s.recordBlockAt(at(2*time.Second), at(500*time.Millisecond))
	// productionInterval = 2s, receiveInterval = 500ms.
	// 500ms < 2s → not caught up.

	require.False(t, s.IsCaughtUpAt(at(500*time.Millisecond)))
}

func TestSyncTrackerNotCatchingUp(t *testing.T) {
	// Receiving slow enough to trigger the rate check (3× production).
	s := NewSyncTracker(10 * time.Second)

	// 4 blocks: production intervals = 2s each, avg = 2s.
	// rate = 3 × 2s = 6s.
	s.recordBlockAt(at(0), at(0))
	s.recordBlockAt(at(2*time.Second), at(500*time.Millisecond))
	s.recordBlockAt(at(4*time.Second), at(1*time.Second))
	s.recordBlockAt(at(6*time.Second), at(2*time.Second))

	// At T=8s: receiveInterval = 6s >= prod*3 = 6s → escape by rate.
	require.True(t, s.IsCaughtUpAt(at(8*time.Second)))
}

func TestSyncTrackerTimeoutFloorAfterProduction(t *testing.T) {
	// Even with production info, the timeout floor still applies.
	// production = 2s, timeout = 1s, receive = 1.5s.
	// 1.5s < 2s → not caught up by rate.
	// 1.5s >= 1s → caught up by timeout floor.
	s := NewSyncTracker(1 * time.Second)

	s.recordBlockAt(at(0), at(0))
	s.recordBlockAt(at(2*time.Second), at(500*time.Millisecond))

	// After 1.5s: rate check not triggered (1.5 < 2), but floor fires (1.5 >= 1).
	require.True(t, s.IsCaughtUpAt(at(2*time.Second)))
}

func TestSyncTrackerZeroTimeout(t *testing.T) {
	// Zero timeout: caught up immediately (should not happen in practice).
	s := NewSyncTracker(0)

	s.recordBlockAt(at(0), at(0))

	require.True(t, s.IsCaughtUpAt(at(0)))
}

func TestSyncTrackerThrottling(t *testing.T) {
	// Attacker slows delivery to match production rate.
	s := NewSyncTracker(5 * time.Second)

	// Fast catch-up phase: blocks arrive faster than production.
	s.recordBlockAt(at(0), at(0))
	for i := 0; i < 3; i++ {
		s.recordBlockAt(at(time.Duration(2+i*2)*time.Second), at(time.Duration(100+i*100)*time.Millisecond))
	}
	// production avg ≈ 2s, last receive 100ms. Not caught up.
	require.False(t, s.IsCaughtUpAt(at(500*time.Millisecond)))

	// Attacker throttles: production ≈ 2.5s, wall at 3s.
	// rate: 3 × prod = 7.5s. At T=8s: receiveInterval = 5s >= timeout floor → escape.
	s.recordBlockAt(at(10*time.Second), at(3*time.Second))
	require.True(t, s.IsCaughtUpAt(at(8*time.Second)),
		"attacker throttling to production rate: escape")
}
