package consensus

import (
	"errors"
	"fmt"
	"reflect"
)

var (
	ErrNilMessage                    = errors.New("message is nil")
	ErrPeerStateHeightRegression     = errors.New("error peer state height regression")
	ErrPeerStateInvalidStartTime     = errors.New("error peer state invalid startTime")
	ErrCommitQuorumNotMet            = errors.New("extended commit does not have +2/3 majority")
	ErrNilPrivValidator              = errors.New("entered createProposalBlock with privValidator being nil")
	ErrProposalWithoutPreviousCommit = errors.New("propose step; cannot propose anything without commit for the previous block")
)

// Consensus sentinel errors
var (
	ErrInvalidProposalSignature   = errors.New("error invalid proposal signature")
	ErrInvalidProposalPOLRound    = errors.New("error invalid proposal POL round")
	ErrAddingVote                 = errors.New("error adding vote")
	ErrSignatureFoundInPastBlocks = errors.New("found signature from the same key")

	errPubKeyIsNotSet = errors.New("pubkey is not set. Look for \"Can't get private validator pubkey\" errors")
)

type ErrMsgToProto struct {
	MessageName string
	Err         error
}

func NewErrMsgToProto[T any](msg T, err error) ErrMsgToProto {
	t := reflect.TypeOf(msg)

	// if msg provided is a pointer, get the underlying value
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	messageName := t.Name()
	return ErrMsgToProto{MessageName: messageName, Err: err}
}

func (e ErrMsgToProto) Error() string {
	return fmt.Sprintf("%s msg to proto error: %s", e.MessageName, e.Err.Error())
}

func (e ErrMsgToProto) Unwrap() error {
	return e.Err
}
