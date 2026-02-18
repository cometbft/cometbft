package consensus

import (
	types "github.com/cometbft/cometbft/types"
	"github.com/pkg/errors"
)

// IngestVerifiedBlock executed and commits a verified block into the consensus state.
// If the verifier determined that this block has vote extensions enabled, extCommit is not nil
// (mutually exclusive with commit)
func (cs *State) IngestVerifiedBlock(
	block *types.Block,
	blockParts *types.PartSet,
	commit *types.Commit,
	extCommit *types.ExtendedCommit,
) (err error, malicious bool) {
	switch {
	case block == nil:
		return errors.Wrap(ErrValidation, "block is nil"), false
	case blockParts == nil:
		return errors.Wrap(ErrValidation, "part set is nil"), false
	case commit == nil && extCommit == nil:
		return errors.Wrap(ErrValidation, "commit and extCommit are both nil"), false
	case commit != nil && extCommit != nil:
		return errors.Wrap(ErrValidation, "commit and extCommit are both not nil"), false
	}

	var (
		height            = block.Height
		extensionsEnabled = extCommit != nil
		blockID           = types.BlockID{
			Hash:          block.Hash(),
			PartSetHeader: blockParts.Header(),
		}
	)

	cs.mtx.Lock()
	defer cs.mtx.Unlock()

	var (
		stateCopy = cs.state.Copy()
		logger    = cs.Logger.With("height", height)
	)

	if height <= stateCopy.LastBlockHeight {
		logger.Debug("Block already included")
		return ErrAlreadyIncluded, false
	}

	if height != stateCopy.LastBlockHeight+1 {
		return errors.Wrapf(ErrHeightGap, "got %d, want %d", height, stateCopy.LastBlockHeight+1), false
	}

	// this is not thread-safe, thus we must exec it under the lock
	if err := cs.blockExec.ValidateBlock(stateCopy, block); err != nil {
		return errors.Wrap(err, "failed to validate block"), true
	}

	if extensionsEnabled {
		cs.blockStore.SaveBlockWithExtendedCommit(block, blockParts, extCommit)
	} else {
		cs.blockStore.SaveBlock(block, blockParts, commit)
	}

	// NOTE: fsync
	if err := cs.wal.WriteSync(EndHeightMessage{height}); err != nil {
		panic(errors.Wrapf(err, "unable to write end height message to WAL for height %d", height))
	}

	// the follow flow is similar to finalizeCommit(height)
	stateCopy, err = cs.blockExec.ApplyVerifiedBlock(stateCopy, blockID, block)
	if err != nil {
		// we can't recover from this error, so we panic
		panic(errors.Wrapf(err, "failed to apply verified block (height: %d, hash: %x)", block.Height, block.Hash()))
	}

	// must be called before we update state
	cs.recordMetrics(height, block)

	// NewHeightStep!
	cs.updateToState(stateCopy)

	// Private validator might have changed it's key pair => refetch pubkey.
	if err := cs.updatePrivValidatorPubKey(); err != nil {
		logger.Error("Failed to get private validator pubkey", "err", err)
	}

	// todo: what would happen if we have 100s of ingested block like this in a row?
	cs.scheduleRound0(&cs.RoundState)

	return nil, false
}
