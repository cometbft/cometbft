package types

import (
	"bytes"
	"errors"
	"fmt"

	"github.com/cometbft/cometbft/proto/tendermint/privval"

	"github.com/cometbft/cometbft/crypto"
	"github.com/cometbft/cometbft/crypto/ed25519"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
)

// PrivValidator defines the functionality of a local CometBFT validator
// that signs votes and proposals, and never double signs.
type PrivValidator interface {
	GetPubKey() (crypto.PubKey, error)

	SignVote(chainID string, vote *cmtproto.Vote) error
	SignProposal(chainID string, proposal *cmtproto.Proposal) error
	SignRawBytes(chainID, uniqueID string, rawBytes []byte) ([]byte, error)
}

// RawBytesSignBytesPrefix defines a domain separator prefix added to raw bytes to ensure the resulting
// signed message can't be confused with a consensus message, which could lead to double signing
const RawBytesSignBytesPrefix = "COMET::RAW_BYTES::SIGN"

// RawBytesMessageSignBytes returns the canonical bytes for signing raw data messages.
// It requires non-empty chainID, uniqueID, and rawBytes to prevent security issues.
// Returns error if any required parameter is empty or if marshaling fails.
func RawBytesMessageSignBytes(chainID, uniqueID string, rawBytes []byte) ([]byte, error) {
	if chainID == "" {
		return nil, errors.New("chainID cannot be empty")
	}

	if uniqueID == "" {
		return nil, fmt.Errorf("uniqueID cannot be empty")
	}

	if len(rawBytes) == 0 {
		return nil, fmt.Errorf("rawBytes cannot be empty")
	}

	prefix := []byte(RawBytesSignBytesPrefix)

	signRequest := &privval.SignRawBytesRequest{
		ChainId:  chainID,
		RawBytes: rawBytes,
		UniqueId: uniqueID,
	}
	protoBytes, err := signRequest.Marshal()
	if err != nil {
		return nil, err
	}
	return append(prefix, protoBytes...), nil
}

type PrivValidatorsByAddress []PrivValidator

func (pvs PrivValidatorsByAddress) Len() int {
	return len(pvs)
}

func (pvs PrivValidatorsByAddress) Less(i, j int) bool {
	pvi, err := pvs[i].GetPubKey()
	if err != nil {
		panic(err)
	}
	pvj, err := pvs[j].GetPubKey()
	if err != nil {
		panic(err)
	}

	return bytes.Compare(pvi.Address(), pvj.Address()) == -1
}

func (pvs PrivValidatorsByAddress) Swap(i, j int) {
	pvs[i], pvs[j] = pvs[j], pvs[i]
}

//----------------------------------------
// MockPV

// MockPV implements PrivValidator without any safety or persistence.
// Only use it for testing.
type MockPV struct {
	PrivKey              crypto.PrivKey
	breakProposalSigning bool
	breakVoteSigning     bool
}

var _ PrivValidator = &MockPV{}

func NewMockPV() MockPV {
	return MockPV{ed25519.GenPrivKey(), false, false}
}

// NewMockPVWithParams allows one to create a MockPV instance, but with finer
// grained control over the operation of the mock validator. This is useful for
// mocking test failures.
func NewMockPVWithParams(privKey crypto.PrivKey, breakProposalSigning, breakVoteSigning bool) MockPV {
	return MockPV{privKey, breakProposalSigning, breakVoteSigning}
}

// Implements PrivValidator.
func (pv MockPV) GetPubKey() (crypto.PubKey, error) {
	return pv.PrivKey.PubKey(), nil
}

// Implements PrivValidator.
func (pv MockPV) SignVote(chainID string, vote *cmtproto.Vote) error {
	useChainID := chainID
	if pv.breakVoteSigning {
		useChainID = "incorrect-chain-id"
	}

	signBytes := VoteSignBytes(useChainID, vote)
	sig, err := pv.PrivKey.Sign(signBytes)
	if err != nil {
		return err
	}
	vote.Signature = sig

	var extSig []byte
	// We only sign vote extensions for non-nil precommits
	if vote.Type == cmtproto.PrecommitType && !ProtoBlockIDIsNil(&vote.BlockID) {
		extSignBytes := VoteExtensionSignBytes(useChainID, vote)
		extSig, err = pv.PrivKey.Sign(extSignBytes)
		if err != nil {
			return err
		}
	} else if len(vote.Extension) > 0 {
		return errors.New("unexpected vote extension - vote extensions are only allowed in non-nil precommits")
	}
	vote.ExtensionSignature = extSig
	return nil
}

func (pv MockPV) SignRawBytes(chainID, uniqueID string, rawBytes []byte) ([]byte, error) {
	useChainID := chainID
	if pv.breakProposalSigning {
		useChainID = "incorrect-chain-id"
	}

	signBytes, err := RawBytesMessageSignBytes(useChainID, uniqueID, rawBytes)
	if err != nil {
		return nil, err
	}
	sig, err := pv.PrivKey.Sign(signBytes)
	if err != nil {
		return nil, err
	}
	return sig, nil
}

// Implements PrivValidator.
func (pv MockPV) SignProposal(chainID string, proposal *cmtproto.Proposal) error {
	useChainID := chainID
	if pv.breakProposalSigning {
		useChainID = "incorrect-chain-id"
	}

	signBytes := ProposalSignBytes(useChainID, proposal)
	sig, err := pv.PrivKey.Sign(signBytes)
	if err != nil {
		return err
	}
	proposal.Signature = sig
	return nil
}

func (pv MockPV) ExtractIntoValidator(votingPower int64) *Validator {
	pubKey, _ := pv.GetPubKey()
	return &Validator{
		Address:     pubKey.Address(),
		PubKey:      pubKey,
		VotingPower: votingPower,
	}
}

// String returns a string representation of the MockPV.
func (pv MockPV) String() string {
	mpv, _ := pv.GetPubKey() // mockPV will never return an error, ignored here
	return fmt.Sprintf("MockPV{%v}", mpv.Address())
}

// XXX: Implement.
func (pv MockPV) DisableChecks() {
	// Currently this does nothing,
	// as MockPV has no safety checks at all.
}

type ErroringMockPV struct {
	MockPV
}

var ErroringMockPVErr = errors.New("erroringMockPV always returns an error")

// Implements PrivValidator.
func (pv *ErroringMockPV) SignVote(string, *cmtproto.Vote) error {
	return ErroringMockPVErr
}

// Implements PrivValidator.
func (pv *ErroringMockPV) SignProposal(string, *cmtproto.Proposal) error {
	return ErroringMockPVErr
}

// NewErroringMockPV returns a MockPV that fails on each signing request. Again, for testing only.

func NewErroringMockPV() *ErroringMockPV {
	return &ErroringMockPV{MockPV{ed25519.GenPrivKey(), false, false}}
}
