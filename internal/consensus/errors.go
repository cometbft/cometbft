package consensus

import (
	"errors"
	"fmt"
)

var (
	ErrNilMessage                    = errors.New("message is nil")
	ErrPeerStateHeightRegression     = errors.New("error peer state height regression")
	ErrPeerStateInvalidStartTime     = errors.New("error peer state invalid startTime")
	ErrCommitQuorumNotMet            = errors.New("extended commit does not have +2/3 majority")
	ErrNilPrivValidator              = errors.New("entered createProposalBlock with privValidator being nil")
	ErrProposalWithoutPreviousCommit = errors.New("propose step; cannot propose anything without commit for the previous block")
)

// Consensus sentinel errors.
var (
	ErrInvalidProposalSignature   = errors.New("error invalid proposal signature")
	ErrInvalidProposalPOLRound    = errors.New("error invalid proposal POL round")
	ErrAddingVote                 = errors.New("error adding vote")
	ErrSignatureFoundInPastBlocks = errors.New("found signature from the same key")
	ErrPubKeyIsNotSet             = errors.New("pubkey is not set. Look for \"Can't get private validator pubkey\" errors")
)

type ErrConsensusMessageNotRecognized struct {
	Message any
}

func (e ErrConsensusMessageNotRecognized) Error() string {
	return fmt.Sprintf("consensus: message not recognized: %T", e.Message)
}

type ErrDenyMessageOverflow struct {
	Err error
}

func (e ErrDenyMessageOverflow) Error() string {
	return fmt.Sprintf("denying message due to possible overflow: %s", e.Err.Error())
}

func (e ErrDenyMessageOverflow) Unwrap() error {
	return e.Err
}
