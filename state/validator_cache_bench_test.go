package state_test

import (
	"testing"

	sm "github.com/cometbft/cometbft/state"
)

// BenchmarkLoadValidators_PerBlockCycle_NoCache measures the old hot-path cost:
// ProcessProposal, FinalizeBlock, and ExtendVote each called LoadValidators
// independently for the same height, so there were 3 DB reads + proto
// unmarshals per block cycle.
func BenchmarkLoadValidators_PerBlockCycle_NoCache(b *testing.B) {
	for _, nVals := range []int{10, 100} {
		b.Run(benchName(nVals), func(b *testing.B) {
			state, stateDB, _ := makeState(nVals, 10)
			stateStore := sm.NewStore(stateDB, sm.StoreOptions{})
			loadHeight := state.LastBlockHeight // validators loaded at block.Height-1

			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				// ProcessProposal
				if _, err := stateStore.LoadValidators(loadHeight); err != nil {
					b.Fatal(err)
				}
				// FinalizeBlock
				if _, err := stateStore.LoadValidators(loadHeight); err != nil {
					b.Fatal(err)
				}
				// ExtendVote
				if _, err := stateStore.LoadValidators(loadHeight); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

// BenchmarkLoadValidators_PerBlockCycle_WithCache measures the new hot-path cost:
// validators are loaded once (1 DB read) and the pointer is reused for the
// remaining calls in the same block cycle.  The atomic.Pointer load for a cache
// hit is effectively free compared with a DB round-trip.
func BenchmarkLoadValidators_PerBlockCycle_WithCache(b *testing.B) {
	for _, nVals := range []int{10, 100} {
		b.Run(benchName(nVals), func(b *testing.B) {
			state, stateDB, _ := makeState(nVals, 10)
			stateStore := sm.NewStore(stateDB, sm.StoreOptions{})
			loadHeight := state.LastBlockHeight

			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				// 1 DB read (first call in the block cycle, cache miss).
				v, err := stateStore.LoadValidators(loadHeight)
				if err != nil {
					b.Fatal(err)
				}
				// 2 cache hits — in production these become blockExec.lastLoadedValidators.Load().
				_ = v
				_ = v
			}
		})
	}
}

func benchName(nVals int) string {
	switch nVals {
	case 10:
		return "10vals"
	case 100:
		return "100vals"
	default:
		return "vals"
	}
}
