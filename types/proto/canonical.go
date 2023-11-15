package types

import (
	"github.com/cometbft/cometbft/api/cometbft/types/v1beta1"
	"github.com/cometbft/cometbft/api/cometbft/types/v1beta3"
)

type (
	CanonicalBlockID       = v1beta1.CanonicalBlockID
	CanonicalPartSetHeader = v1beta1.CanonicalPartSetHeader
	CanonicalProposal      = v1beta1.CanonicalProposal
	CanonicalVote          = v1beta1.CanonicalVote
	CanonicalVoteExtension = v1beta3.CanonicalVoteExtension
)
