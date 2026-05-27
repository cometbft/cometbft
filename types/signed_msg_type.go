package types

import cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"

// Short lowercase descriptors for SignedMsgType values; exported so tests and
// callers can reuse them.
const (
	UnknownShortName   = "unknown"
	PrevoteShortName   = "prevote"
	PrecommitShortName = "precommit"
	ProposalShortName  = "proposal"
)

// IsVoteTypeValid returns true if t is a valid vote type.
func IsVoteTypeValid(t cmtproto.SignedMsgType) bool {
	switch t {
	case cmtproto.PrevoteType, cmtproto.PrecommitType:
		return true
	default:
		return false
	}
}

var signedMsgTypeToShortName = map[cmtproto.SignedMsgType]string{
	cmtproto.UnknownType:   UnknownShortName,
	cmtproto.PrevoteType:   PrevoteShortName,
	cmtproto.PrecommitType: PrecommitShortName,
	cmtproto.ProposalType:  ProposalShortName,
}

// Returns a short lowercase descriptor for a signed message type.
func SignedMsgTypeToShortString(t cmtproto.SignedMsgType) string {
	if shortName, ok := signedMsgTypeToShortName[t]; ok {
		return shortName
	}
	return UnknownShortName
}
