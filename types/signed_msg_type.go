package types

import cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"

// IsVoteTypeValid returns true if t is a valid vote type.
func IsVoteTypeValid(t cmtproto.SignedMsgType) bool {
	switch t {
	case cmtproto.PrevoteType, cmtproto.PrecommitType:
		return true
	default:
		return false
	}
}

var signedMsgTypeToShortName = map[SignedMsgType]string{
	UnknownType:   "unknown",
	PrevoteType:   "prevote",
	PrecommitType: "precommit",
	ProposalType:  "proposal",
}

// Returns a short lowercase descriptor for a signed message type.
func SignedMsgTypeToShortString(t SignedMsgType) string {
	if shortName, ok := signedMsgTypeToShortName[t]; ok {
		return shortName
	}
	return "unknown"
}
