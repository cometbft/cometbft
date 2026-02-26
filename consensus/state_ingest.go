package consensus

import (
	"fmt"
	"time"

	cstypes "github.com/cometbft/cometbft/consensus/types"
	"github.com/cometbft/cometbft/state"
	types "github.com/cometbft/cometbft/types"
	cmttime "github.com/cometbft/cometbft/types/time"
	"github.com/pkg/errors"
)

// IngestCandidate is a block that *might* be ingested into the consensus state. Not thread safe.
type IngestCandidate struct {
	block      *types.Block
	blockParts *types.PartSet
	commit     *types.Commit
	extCommit  *types.ExtendedCommit

	verified bool

	// caches IngestCandidate.BlockID() to avoid recalculating it
	cachedBlockID types.BlockID
}

type ingestVerifiedBlockRequest struct {
	IngestCandidate
	sentAt   time.Time
	response chan ingestVerifiedBlockResponse
}

type ingestVerifiedBlockResponse struct {
	err error
}

// NewIngestCandidate constructs IngestCandidate.
func NewIngestCandidate(
	block *types.Block,
	blockParts *types.PartSet,
	commit *types.Commit,
	extCommit *types.ExtendedCommit,
) (IngestCandidate, error) {
	ic := IngestCandidate{
		block:      block,
		blockParts: blockParts,
		commit:     commit,
		extCommit:  extCommit,
	}

	if err := ic.ValidateBasic(); err != nil {
		return ic, err
	}

	return ic, nil
}

// Height returns the height of the block.
func (ic *IngestCandidate) Height() int64 {
	return ic.block.Height
}

// BlockID returns the block ID
func (ic *IngestCandidate) BlockID() types.BlockID {
	if !ic.cachedBlockID.IsZero() {
		return ic.cachedBlockID
	}

	ic.cachedBlockID = types.BlockID{
		Hash:          ic.block.Hash(),
		PartSetHeader: ic.blockParts.Header(),
	}

	return ic.cachedBlockID
}

// ValidateBasic validates the ingest candidate.
func (ic *IngestCandidate) ValidateBasic() error {
	switch {
	case ic.block == nil:
		return errors.Wrap(ErrValidation, "block is nil")
	case ic.blockParts == nil:
		return errors.Wrap(ErrValidation, "part set is nil")
	case ic.commit == nil:
		return errors.Wrap(ErrValidation, "commit is nil")
	}

	// validate commit/extCommit
	var (
		blockID     = ic.BlockID()
		blockHeight = ic.block.Height
	)

	if ic.extensionsEnabled() {
		switch {
		case ic.extCommit.Height != blockHeight:
			return errors.Wrapf(ErrValidation, "extCommit height mismatch: got %d, want %d", ic.extCommit.Height, blockHeight)
		case !ic.extCommit.BlockID.Equals(blockID):
			return errors.Wrap(ErrValidation, "extended commit blockID mismatch")
		default:
			return ic.extCommit.ValidateBasic()
		}
	}

	switch {
	case ic.commit.Height != blockHeight:
		return errors.Wrapf(ErrValidation, "commit height mismatch: got %d, want %d", ic.commit.Height, blockHeight)
	case !ic.commit.BlockID.Equals(blockID):
		return errors.Wrap(ErrValidation, "commit blockID mismatch")
	default:
		return ic.commit.ValidateBasic()
	}
}

// Verify verifies the block against provided state using light client verification.
func (ic *IngestCandidate) Verify(state state.State) error {
	var (
		height            = ic.Height()
		blockID           = ic.BlockID()
		chainID           = state.ChainID
		extensionsPresent = state.ConsensusParams.ABCI.VoteExtensionsEnabled(height)
	)

	// ensure invariant
	if extensionsPresent != ic.extensionsEnabled() {
		return fmt.Errorf(
			"invalid ext commit state: height %d: extensionsPresent=%t, extensionsEnabled=%t",
			ic.Height(), extensionsPresent, ic.extensionsEnabled(),
		)
	}

	if err := state.ValidateBlock(ic.block); err != nil {
		return fmt.Errorf("validate block: %w", err)
	}

	// verify commit
	err := state.Validators.VerifyCommitLight(chainID, blockID, height, ic.commit)
	if err != nil {
		return fmt.Errorf("verify commit: %w", err)
	}

	// verify commit extensions
	if ic.extensionsEnabled() {
		if err = ic.extCommit.EnsureExtensions(true); err != nil {
			return fmt.Errorf("ensure extensions: %w", err)
		}

		err = state.Validators.VerifyCommitLight(chainID, blockID, height, ic.extCommit.ToCommit())
		if err != nil {
			return fmt.Errorf("verify extended commit: %w", err)
		}
	}

	ic.verified = true

	return nil
}

func (ic *IngestCandidate) extensionsEnabled() bool {
	return ic.extCommit != nil
}

// commitVoting returns the commit round and vote set for the verified block.
func (ic *IngestCandidate) commitVoting(chainID string, vals *types.ValidatorSet) (round int32, voteSet *types.VoteSet) {
	if ic.extensionsEnabled() {
		return ic.extCommit.Round, ic.extCommit.ToExtendedVoteSet(chainID, vals)
	}

	return ic.commit.Round, ic.commit.ToVoteSet(chainID, vals)
}

// IngestVerifiedBlock ingests a next VERIFIED valid block into the consensus state.
// Verification is the domain responsibility of the caller (otherwise the consensus will panic).
// It uses the underlying internalQueue to ensure SERIAL state-machine processing inside the main receiveRoutine.
// See handleIngestVerifiedBlock for the actual implementation and error handling.
func (cs *State) IngestVerifiedBlock(ic IngestCandidate) (err error) {
	logger := cs.Logger.With("height", ic.Height())
	logger.Info("ingesting verified block")

	defer func() {
		if err != nil {
			logger.Info("failed to ingest verified block", "err", err)
		} else {
			logger.Info("ingested verified block")
		}
	}()

	// register response channel so we can receive from receiveRoutine
	ch := make(chan ingestVerifiedBlockResponse, 1)
	defer close(ch)

	req := &ingestVerifiedBlockRequest{
		IngestCandidate: ic,
		sentAt:          time.Now(),
		response:        ch,
	}

	cs.sendInternalMessage(msgInfo{Msg: req})

	select {
	case <-cs.Quit():
		return fmt.Errorf("consensus shutdown")
	case res := <-req.response:
		return res.err
	}
}

// note the outcome of this call is NOT relevant to statsMsgQueue
func (cs *State) handleIngestVerifiedBlockRequest(req *ingestVerifiedBlockRequest) {
	err := cs.ingestBlock(req.IngestCandidate)

	req.response <- ingestVerifiedBlockResponse{err: err}
}

// ingestBlock handles the ingestion of IngestCandidate into the consensus.
// Note that the MUTEX is held by the caller and IngestCandidate should be already validated&verified.
// Might return ErrAlreadyIncluded, ErrHeightGap, or ErrValidation, or other errors.
func (cs *State) ingestBlock(ic IngestCandidate) error {
	if !ic.verified {
		return errors.Wrap(ErrValidation, "unverified ingest candidate")
	}

	var (
		block           = ic.block
		blockParts      = ic.blockParts
		height          = ic.block.Height
		lastBlockHeight = cs.state.LastBlockHeight
	)

	// olb block or consensus already progressed
	// between IngestVerifiedBlock() and ingestBlock() calls.
	if height <= lastBlockHeight {
		return ErrAlreadyIncluded
	}

	// a future block submitted eg (state=100, ingestCandidate=105)
	// it's up to the caller to handle this error.
	if height != lastBlockHeight+1 {
		return errors.Wrapf(ErrHeightGap, "got %d, want %d", height, lastBlockHeight+1)
	}

	// okay, at this point this is definitely the next valid and verified block, because
	// ingestBlock is called in serial from the main receiveRoutine.
	var (
		stateCopy = cs.state.Copy()
		logger    = cs.Logger.With("height", height)
	)

	// ============ enterCommit(height, commitRound) ============
	commitRound, commitVoteSet := ic.commitVoting(stateCopy.ChainID, stateCopy.Validators)

	cs.updateRoundStep(commitRound, cstypes.RoundStepCommit)
	cs.CommitRound = commitRound
	cs.LastCommit = commitVoteSet
	cs.CommitTime = cmttime.Now()
	cs.newStep()

	// ============ finalizeCommit(height) ============

	// this will also update blockStore.Height,
	// so blocksync responds to peers with the correct height.
	if ic.extensionsEnabled() {
		cs.blockStore.SaveBlockWithExtendedCommit(block, blockParts, ic.extCommit)
	} else {
		cs.blockStore.SaveBlock(block, blockParts, ic.commit)
	}

	// NOTE: fsync
	if err := cs.wal.WriteSync(EndHeightMessage{height}); err != nil {
		panic(errors.Wrapf(err, "unable to write end height message to WAL for height %d", height))
	}

	// the following flow is similar to finalizeCommit(height)
	stateCopy, err := cs.blockExec.ApplyVerifiedBlock(stateCopy, ic.BlockID(), block)
	if err != nil {
		// we can't recover from this error
		panic(errors.Wrapf(err, "failed to apply verified block (height: %d, hash: %x)", block.Height, block.Hash()))
	}

	// must be called before we update state
	cs.recordMetrics(height, block)

	// NewHeightStep!
	// - The block is already verified using the light client
	// - cs.LastCommit is set and valid
	// - cs.Votes contain votes that are NOT relevant to this block
	// - drop votes to avoid updateToState() 2/3 majority check (`if cs.CommitRound > -1 && cs.Votes != nil {...}`)
	cs.Votes = nil
	cs.updateToState(stateCopy)

	// private validator might have changed its key pair => refetch pubkey.
	if err := cs.updatePrivValidatorPubKey(); err != nil {
		logger.Error("Failed to get private validator pubkey", "err", err)
	}

	cs.scheduleRound0(&cs.RoundState)

	return nil
}
