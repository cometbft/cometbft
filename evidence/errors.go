package evidence

import (
	"fmt"

	"github.com/cometbft/cometbft/libs/bytes"
	"github.com/cometbft/cometbft/types"
)

type (
	ErrNoHeaderAtHeight struct {
		Height int64
	}

	ErrNoCommitAtHeight struct {
		Height int64
	}

	ErrUnrecognizedEvidenceType struct {
		Evidence types.Evidence
	}

	// ErrVotingPowerDoesNotMatch is returned when voting power from trusted validator set does not match voting power from evidence
	ErrVotingPowerDoesNotMatch struct {
		TrustedVotingPower  int64
		EvidenceVotingPower int64
	}

	ErrAddressNotValidatorAtHeight struct {
		Address bytes.HexBytes
		Height  int64
	}

	// ErrValidatorAddressesDoNotMatch is returned when provided DuplicateVoteEvidence's votes have different validators as signers
	ErrValidatorAddressesDoNotMatch struct {
		ValidatorA bytes.HexBytes
		ValidatorB bytes.HexBytes
	}
)

// func (e ) Error() string {
// 	return fmt.Sprintf("",)
// }

func (e ErrNoHeaderAtHeight) Error() string {
	return fmt.Sprintf("don't have header at height #%d", e.Height)
}

func (e ErrNoCommitAtHeight) Error() string {
	return fmt.Sprintf("don't have commit at height #%d", e.Height)
}

func (e ErrUnrecognizedEvidenceType) Error() string {
	return fmt.Sprintf("unrecognized evidence type: %T", e.Evidence)
}

func (e ErrVotingPowerDoesNotMatch) Error() string {
	return fmt.Sprintf("total voting power from the evidence and our validator set does not match (%d != %d)", e.TrustedVotingPower, e.EvidenceVotingPower)
}

func (e ErrAddressNotValidatorAtHeight) Error() string {
	return fmt.Sprintf("address %X was not a validator at height %d", e.Address, e.Height)
}

func (e ErrValidatorAddressesDoNotMatch) Error() string {
	return fmt.Sprintf("validator addresses do not match: %X vs %X",
		e.ValidatorA,
		e.ValidatorB,
	)
}
