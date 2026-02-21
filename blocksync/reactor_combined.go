package blocksync

import (
	"fmt"
	"time"

	// "github.com/cometbft/cometbft/consensus"
	"github.com/cometbft/cometbft/consensus"
	"github.com/cometbft/cometbft/types"
	"github.com/pkg/errors"
)

// BlockIngestor represents a reactor that can ingest blocks into the consensus state.
type BlockIngestor interface {
	IngestVerifiedBlock(block consensus.VerifiedBlock) (err error, malicious bool)
}

func (r *Reactor) getBlockIngestor() (BlockIngestor, error) {
	cr, ok := r.Switch.Reactor("CONSENSUS")
	if !ok {
		return nil, errors.New("consensus reactor not found")
	}

	bi, ok := cr.(BlockIngestor)
	if !ok {
		return nil, errors.Errorf("reactor %T does not implement BlockIngestor", cr)
	}

	return bi, nil
}

// blockIngestorRoutine is a similar loop to poolRoutine, but for combined mode.
// It fetches consecutive blocks from the pool, ensures invariants, performs commit validation using
// the light client and then passes it to BlockIngestor (consensus).
// Influence on networking and block sharing: as consensus and blocksync reactors both point to the same BlockStore,
// blocksync req/res always operate on the latest state --> no need to explicitly update blocksync's state
func (r *Reactor) blockIngestorRoutine(blockIngestor BlockIngestor) {
	r.Logger.Info("Starting blocksync pool routine (combined mode)")

	trySyncTicker := time.NewTicker(intervalTrySync)
	defer trySyncTicker.Stop()

	syncIterationCh := make(chan struct{}, 1)
	defer close(syncIterationCh)

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
			// See if there are any blocks to sync. We need two consecutive blocks
			// in order to perform blocksync verification.
			blockA, blockB, extCommitA := r.pool.PeekTwoBlocks()
			if blockA == nil || blockB == nil {
				continue FOR_LOOP
			}

			// sanity check
			if blockA.Height+1 != blockB.Height {
				panic(fmt.Errorf(
					"heights of first and second block are not consecutive (want %d, got %d)",
					blockA.Height+1,
					blockB.Height,
				))
			}

			// note this is a db call. consider caching in mem.
			// right now it's the safest and the easiest & safest way to fetch the latest state.
			state, err := r.blockExec.Store().Load()
			if err != nil {
				r.Logger.Error("Failed to load latest state. Halting blocksync.", "err", err)
				return
			}

			latestHeight := state.LastBlockHeight

			// this means that CONSENSUS reactor has concurrently processed higher block(s).
			// simply pop block A and continue
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
				// should not happen
				r.Logger.Error("Failed to make part set. Halting blocksync.", "height", blockA.Height, "err", err)
				return
			}

			ida := types.BlockID{
				Hash:          blockA.Hash(),
				PartSetHeader: partsA.Header(),
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

			// note that between state fetch and ingest, the state may have changed
			// concurrently by the consensus.
			err, malicious := blockIngestor.IngestVerifiedBlock(consensus.VerifiedBlock{
				Block:      blockA,
				BlockParts: partsA,
				Commit:     blockB.LastCommit,
				ExtCommit:  extCommitA,
			})

			switch {
			case errors.Is(err, consensus.ErrAlreadyIncluded):
				r.Logger.Info("Block included concurrently. Skipping", "height", blockA.Height)
				r.metrics.AlreadyIncluded.Add(1)
				continue FOR_LOOP
			case err != nil && malicious:
				r.metrics.RejectedBlocks.Add(1)
				r.handleValidationFailure(blockA, blockB, err)
				continue FOR_LOOP
			case err != nil:
				// one of [consensus.ErrValidation, consensus.ErrHeightGap, or other...]
				// this is mostly likely an unrecoverable invariant violation that should not happen
				// or should be considered a bug.
				r.Logger.Error("Failed to ingest verified block. Halting blocksync.", "height", blockA.Height, "err", err)
				return
			default:
				r.metrics.recordBlockMetrics(blockA)
				r.metrics.IngestedBlocks.Add(1)
			}
		}
	}
}
