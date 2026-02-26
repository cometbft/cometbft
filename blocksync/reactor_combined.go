package blocksync

import (
	"errors"
	"fmt"
	"time"

	"github.com/cometbft/cometbft/consensus"
	"github.com/cometbft/cometbft/types"
)

// BlockIngestor represents a reactor that can ingest blocks into the consensus state.
type BlockIngestor interface {
	IngestVerifiedBlock(blockCandidate consensus.IngestCandidate) error
}

func (r *Reactor) getBlockIngestor() (BlockIngestor, error) {
	cr, ok := r.Switch.Reactor("CONSENSUS")
	if !ok {
		return nil, fmt.Errorf("consensus reactor not found")
	}

	bi, ok := cr.(BlockIngestor)
	if !ok {
		return nil, fmt.Errorf("reactor %T does not implement BlockIngestor", cr)
	}

	return bi, nil
}

// blockIngestorRoutine is a similar loop to poolRoutine, but for combined mode.
// It fetches consecutive blocks from the pool, ensures invariants, performs commit validation using
// the light client and then passes it to BlockIngestor (consensus).
// Influence on networking and block sharing: as consensus and blocksync reactors both point to the same BlockStore,
// blocksync req/res always operate on the latest state --> no need to explicitly update blocksync's state
func (r *Reactor) blockIngestorRoutine(blockIngestor BlockIngestor) {
	r.Logger.Info("Starting blocksync block ingestor (combined mode)")

	trySyncTicker := time.NewTicker(intervalTrySync)
	defer trySyncTicker.Stop()

	syncIterationCh := make(chan struct{}, 1)
	defer close(syncIterationCh)

	for {
		select {
		case <-r.Quit():
			return
		case <-r.pool.Quit():
			return
		case <-trySyncTicker.C:
			select {
			case syncIterationCh <- struct{}{}:
			default:
				// do nothing, non-blocking
			}
		case <-syncIterationCh:
			// See if there are any blocks to sync. We need two consecutive blocks
			// in order to perform blocksync verification.
			block, nextBlock, extCommit := r.pool.PeekTwoBlocks()
			if block == nil || nextBlock == nil {
				continue
			}

			// sanity check (block pool guarantees sequential heights)
			if block.Height+1 != nextBlock.Height {
				panic(fmt.Errorf(
					"heights of first and second block are not consecutive (want %d, got %d)",
					block.Height+1,
					nextBlock.Height,
				))
			}

			// note this is a db call. consider caching in mem.
			// right now it's the safest and the easiest way to fetch the latest state.
			state, err := r.blockExec.Store().Load()
			if err != nil {
				r.Logger.Error("Failed to load latest state. Halting blocksync", "err", err)
				return
			}

			latestHeight := state.LastBlockHeight

			// this means that CONSENSUS reactor has concurrently processed higher block(s).
			// simply pop the current block and continue
			if block.Height <= latestHeight {
				r.pool.PopRequest()
				r.metrics.AlreadyIncludedBlocks.Add(1)

				r.Logger.Debug(
					"Consensus already processed this block. Skipping",
					"height", block.Height,
					"latest_height", latestHeight,
				)

				continue
			}

			// it should not be possible for blocksync to propose blocks that a too high
			// example: consensus=95, blocksync[100,101,102,...] --> this should not happen
			if block.Height != latestHeight+1 {
				panic(fmt.Errorf(
					"block height gap invariant violated (got %d, want %d)",
					block.Height,
					latestHeight+1,
				))
			}

			if !r.IsRunning() || !r.pool.IsRunning() {
				return
			}

			// try again quickly next loop.
			syncIterationCh <- struct{}{}

			blockParts, err := block.MakePartSet(types.BlockPartSizeBytes)
			if err != nil {
				// should not happen
				r.Logger.Error("Failed to make part set. Halting blocksync", "height", block.Height, "err", err)
				return
			}

			// create ingest candidate block...
			ic, err := consensus.NewIngestCandidate(block, blockParts, nextBlock.LastCommit, extCommit)
			if err != nil {
				r.handleValidationFailure(block, nextBlock, fmt.Errorf("new ingest candidate: %w", err))
				continue
			}

			// ... and verify it against the state
			if err := ic.Verify(state); err != nil {
				r.handleValidationFailure(block, nextBlock, fmt.Errorf("verify ingest candidate: %w", err))
				continue
			}

			// pops `block`
			r.pool.PopRequest()

			// note that between state fetch and ingest, the state may have changed
			// concurrently by the consensus.
			start := time.Now()
			err = blockIngestor.IngestVerifiedBlock(ic)
			elapsed := time.Since(start)

			switch {
			case errors.Is(err, consensus.ErrAlreadyIncluded):
				r.Logger.Info("Block was included concurrently. Skipping", "height", block.Height)
				r.metrics.AlreadyIncludedBlocks.Add(1)
			case err != nil:
				// one of [consensus.ErrValidation, consensus.ErrHeightGap, or other...]
				// most likely it's an unrecoverable invariant violation that should not happen
				// or should be considered a bug.
				r.Logger.Error("Failed to ingest verified block. Halting blocksync", "height", block.Height, "err", err)
				return
			default:
				r.metrics.recordBlockMetrics(block)
				r.metrics.IngestedBlocks.Add(1)
				r.metrics.IngestedBlockDuration.Observe(elapsed.Seconds())
			}
		}
	}
}
