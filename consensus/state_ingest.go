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
// Not thread safe.
type VerifiedBlock struct {
	Block      *types.Block
	BlockParts *types.PartSet
	Commit     *types.Commit
	ExtCommit  *types.ExtendedCommit

	// cache
	blockID types.BlockID
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

// IngestVerifiedBlock ingests a next valid and VERIFIED block into the consensus state.
// Verification is the domain responsibility of the caller (otherwise the consensus will panic).
// It uses the underlying internalQueue instead to ensure SERIAL state-machine processing inside
// the main receiveRoutine. See handleIngestVerifiedBlock for the actual implementation and error handling.
func (cs *State) IngestVerifiedBlock(vb VerifiedBlock) (err error, malicious bool) {
	start := time.Now()

	logger := cs.Logger.With("height", vb.Block.Height)
	logger.Info("ingesting verified block")

	defer func() {
		duration := time.Since(start)

		if err != nil {
			logger.Info("failed to ingest verified block", "dur", duration, "err", err, "malicious", malicious)
		} else {
			logger.Info("ingested verified block", "dur", duration)
		}
	}()

	if err := vb.ValidateBasic(); err != nil {
		return err, false
	}

	// register response channel so we can receive from receiveRoutine
	ch := make(chan ingestVerifiedBlockResponse, 1)
	defer close(ch)

	req := &ingestVerifiedBlockRequest{
		VerifiedBlock: vb,
		sentAt:        time.Now(),
		response:      ch,
	}

	cs.sendInternalMessage(msgInfo{Msg: req})

	res := <-req.response

	return res.err, res.malicious
}

// note the outcome of this call is NOT relevant to statsMsgQueue
func (cs *State) handleIngestVerifiedBlockRequest(req *ingestVerifiedBlockRequest) {
	err, malicious := cs.handleIngestVerifiedBlock(req.VerifiedBlock)

	req.response <- ingestVerifiedBlockResponse{err: err, malicious: malicious}
}

// handleIngestVerifiedBlock handles the ingestion of a verified block into the consensus state.
// note that the MUTEX is held by the caller and VerifiedBlock should be already validated.
func (cs *State) handleIngestVerifiedBlock(vb VerifiedBlock) (err error, malicious bool) {
	var (
		block           = vb.Block
		blockParts      = vb.BlockParts
		height          = vb.Block.Height
		lastBlockHeight = cs.state.LastBlockHeight
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
	// also, an invalid block should mark the peer as malicious
	if err := cs.blockExec.ValidateBlock(stateCopy, block); err != nil {
		return errors.Wrap(err, "failed to validate block"), true
	}

	// ============ enterCommit(height, commitRound) ============
	commitRound, commitVoteSet := vb.CommitVoting(stateCopy.ChainID, stateCopy.Validators)

	cs.updateRoundStep(commitRound, cstypes.RoundStepCommit)
	cs.CommitRound = commitRound
	cs.LastCommit = commitVoteSet
	cs.CommitTime = cmttime.Now()
	cs.newStep()

	// ============ finalizeCommit(height) ============

	// this will also update blockStore.Height,
	// so blocksync responds to peers with the correct height.
	if vb.ExtensionsEnabled() {
		cs.blockStore.SaveBlockWithExtendedCommit(block, blockParts, vb.ExtCommit)
	} else {
		cs.blockStore.SaveBlock(block, blockParts, vb.Commit)
	}

	// NOTE: fsync
	if err := cs.wal.WriteSync(EndHeightMessage{height}); err != nil {
		panic(errors.Wrapf(err, "unable to write end height message to WAL for height %d", height))
	}

	// the follow flow is similar to finalizeCommit(height)
	stateCopy, err = cs.blockExec.ApplyVerifiedBlock(stateCopy, vb.BlockID(), block)
	if err != nil {
		// we can't recover from this error, so we panic
		panic(errors.Wrapf(err, "failed to apply verified block (height: %d, hash: %x)", block.Height, block.Hash()))
	}

	// must be called before we update state
	cs.recordMetrics(height, block)

	// NewHeightStep!
	// drop votes to avoid updateToState() 2/3 majority check (not relevant here)
	cs.Votes = nil
	cs.updateToState(stateCopy)

	// Private validator might have changed it's key pair => refetch pubkey.
	if err := cs.updatePrivValidatorPubKey(); err != nil {
		logger.Error("Failed to get private validator pubkey", "err", err)
	}

	cs.scheduleRound0(&cs.RoundState)

	return nil, false
}

func (vb *VerifiedBlock) ExtensionsEnabled() bool {
	return vb.ExtCommit != nil
}

func (vb *VerifiedBlock) BlockID() types.BlockID {
	if vb.blockID.IsZero() {
		vb.blockID = types.BlockID{
			Hash:          vb.Block.Hash(),
			PartSetHeader: vb.BlockParts.Header(),
		}
	}

	return vb.blockID
}

// CommitVoting returns the commit round and vote set for the verified block.
func (vb *VerifiedBlock) CommitVoting(chainID string, vals *types.ValidatorSet) (round int32, voteSet *types.VoteSet) {
	if vb.ExtensionsEnabled() {
		return vb.ExtCommit.Round, vb.ExtCommit.ToExtendedVoteSet(chainID, vals)
	}

	return vb.Commit.Round, vb.Commit.ToVoteSet(chainID, vals)
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
		return vb.validateCommit()
	}
}

func (vb *VerifiedBlock) validateCommit() error {
	var (
		blockID     = vb.BlockID()
		blockHeight = vb.Block.Height
	)

	if vb.ExtensionsEnabled() {
		switch {
		case vb.ExtCommit.Height != blockHeight:
			return errors.Wrapf(ErrValidation, "extCommit height mismatch: got %d, want %d", vb.ExtCommit.Height, blockHeight)
		case !vb.ExtCommit.BlockID.Equals(blockID):
			return errors.Wrap(ErrValidation, "extended commit blockID mismatch")
		default:
			return vb.ExtCommit.ValidateBasic()
		}
	}

	switch {
	case vb.Commit.Height != blockHeight:
		return errors.Wrapf(ErrValidation, "commit height mismatch: got %d, want %d", vb.Commit.Height, blockHeight)
	case !vb.Commit.BlockID.Equals(blockID):
		return errors.Wrap(ErrValidation, "commit blockID mismatch")
	default:
		return vb.Commit.ValidateBasic()
	}
}
