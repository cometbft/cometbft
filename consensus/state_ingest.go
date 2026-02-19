package consensus

import (
	"time"

	cstypes "github.com/cometbft/cometbft/consensus/types"
	types "github.com/cometbft/cometbft/types"
	cmttime "github.com/cometbft/cometbft/types/time"
	"github.com/pkg/errors"
)

// VerifiedBlock is a block that has been verified by blocksync.
// Commit and ExtCommit are mutually exclusive based on whether vote extensions are enabled at the block height.
type VerifiedBlock struct {
	Block      *types.Block
	BlockParts *types.PartSet
	Commit     *types.Commit
	ExtCommit  *types.ExtendedCommit
}

type ingestVerifiedBlockRequest struct {
	VerifiedBlock
	sentAt   time.Time
	response chan ingestVerifiedBlockResponse
}

type ingestVerifiedBlockResponse struct {
	err       error
	malicious bool
}

// IngestVerifiedBlock ingests a verified block into the consensus state.
func (cs *State) IngestVerifiedBlock(vb VerifiedBlock) (err error, malicious bool) {
	start := time.Now()

	logger := cs.Logger.With("height", vb.Block.Height)
	logger.Info("ingesting verified block")

	defer func() {
		elapsed := time.Since(start)

		if err != nil {
			logger.Info("failed to ingest verified block", "elapsed", elapsed, "err", err, "malicious", malicious)
			return
		}

		logger.Info("ingested verified block", "elapsed", elapsed)
	}()

	if err := vb.ValidateBasic(); err != nil {
		return err, false
	}

	req := &ingestVerifiedBlockRequest{
		VerifiedBlock: vb,
		sentAt:        time.Now(),
		response:      make(chan ingestVerifiedBlockResponse, 1),
	}

	defer close(req.response)

	// use internal queue to ensure we have NO data races with other consensus state machine operations
	// see handleIngestVerifiedBlockMessage
	cs.sendInternalMessage(msgInfo{Msg: req})

	res := <-req.response

	return res.err, res.malicious
}

func (cs *State) handleIngestVerifiedBlockRequest(req *ingestVerifiedBlockRequest) {
	err, malicious := cs.handleIngestVerifiedBlock(req.VerifiedBlock)

	// todo: what to do with cs.statsMsgQueue?

	req.response <- ingestVerifiedBlockResponse{err: err, malicious: malicious}
}

// handleIngestVerifiedBlock handles the ingestion of a verified block into the consensus state.
// note that the MUTEX is held by the caller and VerifiedBlock should be already validated.
func (cs *State) handleIngestVerifiedBlock(vb VerifiedBlock) (err error, malicious bool) {
	var (
		block           = vb.Block
		height          = vb.Block.Height
		blockParts      = vb.BlockParts
		lastBlockHeight = cs.state.LastBlockHeight
		blockID         = types.BlockID{
			Hash:          vb.Block.Hash(),
			PartSetHeader: vb.BlockParts.Header(),
		}

		commit            = vb.Commit
		extCommit         = vb.ExtCommit
		extensionsEnabled = vb.ExtCommit != nil
	)

	if height <= lastBlockHeight {
		return ErrAlreadyIncluded, false
	}

	if height != lastBlockHeight+1 {
		return errors.Wrapf(ErrHeightGap, "got %d, want %d", height, lastBlockHeight+1), false
	}

	var (
		stateCopy = cs.state.Copy()
		logger    = cs.Logger.With("height", height)
	)

	// this is not thread-safe, thus we must exec it under the lock
	if err := cs.blockExec.ValidateBlock(stateCopy, block); err != nil {
		return errors.Wrap(err, "failed to validate block"), true
	}

	// ============ enterCommit(height, commitRound) ============
	// -1 stands for "N/A". It skips bunch of checks in updateToState()
	// that are not relevant as we assume the block is already validated
	// by the light client in blocksync.
	const round = -1

	cs.updateRoundStep(round, cstypes.RoundStepCommit)
	cs.CommitRound = round
	cs.CommitTime = cmttime.Now()
	cs.newStep()

	// ============ finalizeCommit(height) ============

	// this will also update blockStore.Height,
	// so blocksync responds to peers with the correct height.
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

	cs.scheduleRound0(&cs.RoundState)

	return nil, false
}

func (vb *VerifiedBlock) ValidateBasic() error {
	switch {
	case vb.Block == nil:
		return errors.Wrap(ErrValidation, "block is nil")
	case vb.BlockParts == nil:
		return errors.Wrap(ErrValidation, "part set is nil")
	case vb.Commit == nil && vb.ExtCommit == nil:
		return errors.Wrap(ErrValidation, "commit and extCommit are both nil")
	case vb.Commit != nil && vb.ExtCommit != nil:
		return errors.Wrap(ErrValidation, "commit and extCommit are both not nil")
	default:
		return nil
	}
}
