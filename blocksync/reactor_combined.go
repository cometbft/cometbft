package blocksync

import (
	"errors"
	"fmt"
	"time"

	// "github.com/cometbft/cometbft/consensus"
	"github.com/cometbft/cometbft/consensus"
	sm "github.com/cometbft/cometbft/state"
	"github.com/cometbft/cometbft/types"
)

// BlockIngestor represents a reactor that can ingest blocks into the consensus state.
type BlockIngestor interface {
	// Height returns the current COMMITTED height of the consensus state.
	GetLastHeight() int64

	// GetState returns the current consensus state
	GetState() sm.State

	// IngestVerifiedBlock ingests a verified block into the consensus state.
	// commit and extCommit are mutually exclusive based on whether vote extensions are enabled at the block height.
	IngestVerifiedBlock(
		block *types.Block,
		partSet *types.PartSet,
		commit *types.Commit,
		extCommit *types.ExtendedCommit,
	) (err error, malicious bool)
}

// todo: docs
// todo: DRY with poolRoutine
func (r *Reactor) poolCombinedModeRoutine(blockIngestor BlockIngestor) {
	r.Logger.Info("Starting blocksync pool routine (combined mode)")

	trySyncTicker := time.NewTicker(intervalTrySync)
	defer trySyncTicker.Stop()

	syncIterationCh := make(chan struct{}, 1)
	defer func() { close(syncIterationCh) }()

FOR_LOOP:
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
			// See if there are any blocks to sync
			blockA, blockB, extCommitA := r.pool.PeekTwoBlocks()
			if blockA == nil || blockB == nil {
				// we need to have fetched two consecutive blocks in order to
				// perform blocksync verification
				continue FOR_LOOP
			}

			// this means that CONSENSUS reactor has concurrently processed higher block(s).
			// simply pop block A and continue
			latestHeight := blockIngestor.GetLastHeight()

			// sanity check
			if blockA.Height+1 != blockB.Height {
				panic(fmt.Errorf(
					"heights of first and second block are not consecutive (want %d, got %d)",
					blockA.Height+1,
					blockB.Height,
				))
			}

			if blockA.Height <= latestHeight {
				r.pool.PopRequest()
				r.metrics.AlreadyIncluded.Add(1)

				r.Logger.Debug(
					"Consensus already processed this block. Skipping",
					"height", blockA.Height,
					"latest_height", latestHeight,
				)

				continue FOR_LOOP
			}

			if blockA.Height != latestHeight+1 {
				panic(fmt.Errorf(
					"block height gap invariant violated (want %d, got %d)",
					latestHeight+1,
					blockA.Height,
				))
			}

			if !r.IsRunning() || !r.pool.IsRunning() {
				return
			}

			// try again quickly next loop.
			syncIterationCh <- struct{}{}

			partsA, err := blockA.MakePartSet(types.BlockPartSizeBytes)
			if err != nil {
				// todo should we tolerate this error?
				r.Logger.Error("Failed to make part set", "height", blockA.Height, "err", err)
				return
			}

			ida := types.BlockID{
				Hash:          blockA.Hash(),
				PartSetHeader: partsA.Header(),
			}

			state := blockIngestor.GetState()

			// possible early exit: if the current height is not the last block height,
			// concurrent consensus reactor may have already processed a higher block during block parts creation
			if latestHeight != state.LastBlockHeight {
				continue FOR_LOOP
			}

			// verify the first block using the second's commit
			err = state.Validators.VerifyCommitLight(state.ChainID, ida, blockA.Height, blockB.LastCommit)
			if err != nil {
				r.handleValidationFailure(blockA, blockB, err)
				continue FOR_LOOP
			}

			var (
				presentExtCommit  = extCommitA != nil
				extensionsEnabled = state.ConsensusParams.ABCI.VoteExtensionsEnabled(blockA.Height)
			)

			if presentExtCommit != extensionsEnabled {
				err = fmt.Errorf(
					"invalid ext commit state: height %d: presentExtCommit=%t, extensionsEnabled=%t",
					blockA.Height, presentExtCommit, extensionsEnabled,
				)

				r.handleValidationFailure(blockA, blockB, err)
				continue FOR_LOOP
			}

			if extensionsEnabled {
				// if vote extensions were required at this height, ensure they exist.
				if err = extCommitA.EnsureExtensions(true); err != nil {
					r.handleValidationFailure(blockA, blockB, err)
					continue FOR_LOOP
				}
			}

			// pops blockA
			r.pool.PopRequest()

			// note that between state fetch and ingest, the state may have changed concurrently
			// by the consensus reactor
			err, malicious := blockIngestor.IngestVerifiedBlock(blockA, partsA, blockB.LastCommit, extCommitA)

			switch {
			case err != nil && malicious:
				r.metrics.RejectedBlocks.Add(1)
				r.handleValidationFailure(blockA, blockB, err)
				continue FOR_LOOP
			case errors.Is(err, consensus.ErrAlreadyIncluded):
				r.metrics.AlreadyIncluded.Add(1)
				continue FOR_LOOP
			case err != nil:
				// todo figure out how to handle these errors
				r.Logger.Error("Failed to ingest verified block. Halting blocksync.", "height", blockA.Height, "err", err)
				return
			default:
				r.metrics.recordBlockMetrics(blockA)
				r.metrics.IngestedBlocks.Add(1)
			}

			// todo: ensure that pool is aware of recent CONSENSUS height
		}
	}
}

func (r *Reactor) getBlockIngestor() (BlockIngestor, bool) {
	cr, ok := r.Switch.Reactor("CONSENSUS")
	if !ok {
		return nil, false
	}

	crTyped, ok := cr.(*consensus.Reactor)
	if !ok {
		return nil, false
	}

	return crTyped.GetState(), true
}
