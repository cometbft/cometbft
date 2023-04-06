package types

import cmtproto "github.com/cometbft/cometbft/api/cometbft/types/v1"

type SignedMsgType = cmtproto.SignedMsgType

const (
	SignedMsgType_UNKNOWN   SignedMsgType = cmtproto.UnknownType
	SignedMsgType_PREVOTE   SignedMsgType = cmtproto.PrevoteType
	SignedMsgType_PRECOMMIT SignedMsgType = cmtproto.PrecommitType
	SignedMsgType_PROPOSAL  SignedMsgType = cmtproto.ProposalType
)

// IsVoteTypeValid returns true if t is a valid vote type.
func IsVoteTypeValid(t SignedMsgType) bool {
	switch t {
	case SignedMsgType_PREVOTE, SignedMsgType_PRECOMMIT:
		return true
	default:
		return false
	}
}
