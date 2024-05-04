package types

import (
	"github.com/cometbft/cometbft/libs/protoio"
	oracleproto "github.com/cometbft/cometbft/proto/tendermint/oracle"
)

func OracleVoteSignBytes(vote *oracleproto.GossipedVotes) []byte {
	pb := CanonicalizeOracleVote(vote)
	bz, err := protoio.MarshalDelimited(&pb)
	if err != nil {
		panic(err)
	}

	return bz
}

func CanonicalizeOracleVote(vote *oracleproto.GossipedVotes) oracleproto.CanonicalGossipedVotes {
	return oracleproto.CanonicalGossipedVotes{
		Validator:       vote.Validator,
		ValidatorIndex:  vote.ValidatorIndex,
		Votes:           vote.Votes,
		SignedTimestamp: vote.SignedTimestamp,
	}
}
