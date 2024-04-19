package types

import (
	abci "github.com/cometbft/cometbft/abci/types"
	cmtproto "github.com/cometbft/cometbft/api/cometbft/types/v1"
	"github.com/cometbft/cometbft/crypto"
)

// -------------------------------------------------------

// TM2PB is used for converting CometBFT ABCI to protobuf ABCI.
// UNSTABLE.
var TM2PB = tm2pb{}

type tm2pb struct{}

func (tm2pb) Header(header *Header) cmtproto.Header {
	return cmtproto.Header{
		Version: header.Version,
		ChainID: header.ChainID,
		Height:  header.Height,
		Time:    header.Time,

		LastBlockId: header.LastBlockID.ToProto(),

		LastCommitHash: header.LastCommitHash,
		DataHash:       header.DataHash,

		ValidatorsHash:     header.ValidatorsHash,
		NextValidatorsHash: header.NextValidatorsHash,
		ConsensusHash:      header.ConsensusHash,
		AppHash:            header.AppHash,
		LastResultsHash:    header.LastResultsHash,

		EvidenceHash:    header.EvidenceHash,
		ProposerAddress: header.ProposerAddress,
	}
}

func (tm2pb) Validator(val *Validator) abci.Validator {
	return abci.Validator{
		Address: val.PubKey.Address(),
		Power:   val.VotingPower,
	}
}

func (tm2pb) BlockID(blockID BlockID) cmtproto.BlockID {
	return cmtproto.BlockID{
		Hash:          blockID.Hash,
		PartSetHeader: TM2PB.PartSetHeader(blockID.PartSetHeader),
	}
}

func (tm2pb) PartSetHeader(header PartSetHeader) cmtproto.PartSetHeader {
	return cmtproto.PartSetHeader{
		Total: header.Total,
		Hash:  header.Hash,
	}
}

func (tm2pb) ValidatorUpdate(val *Validator) abci.ValidatorUpdate {
	return abci.ValidatorUpdate{
		Power:       val.VotingPower,
		PubKeyBytes: val.PubKey.Bytes(),
		PubKeyType:  val.PubKey.Type(),
	}
}

// XXX: panics on nil or unknown pubkey type.
func (tm2pb) ValidatorUpdates(vals *ValidatorSet) []abci.ValidatorUpdate {
	validators := make([]abci.ValidatorUpdate, vals.Size())
	for i, val := range vals.Validators {
		validators[i] = TM2PB.ValidatorUpdate(val)
	}
	return validators
}

// XXX: panics on nil or unknown pubkey type.
func (tm2pb) NewValidatorUpdate(pubkey crypto.PubKey, power int64) abci.ValidatorUpdate {
	return abci.ValidatorUpdate{
		Power:       power,
		PubKeyBytes: pubkey.Bytes(),
		PubKeyType:  pubkey.Type(),
	}
}

// ----------------------------------------------------------------------------

// PB2TM is used for converting protobuf ABCI to CometBFT ABCI.
// UNSTABLE.
var PB2TM = pb2tm{}

type pb2tm struct{}

func (pb2tm) ValidatorUpdates(vals []abci.ValidatorUpdate) ([]*Validator, error) {
	cmtVals := make([]*Validator, len(vals))
	for i, v := range vals {
		pubKey, err := abci.PubKeyFromValidatorUpdate(v)
		if err != nil {
			return nil, err
		}
		cmtVals[i] = NewValidator(pubKey, v.Power)
	}
	return cmtVals, nil
}
