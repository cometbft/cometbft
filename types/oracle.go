package types

import (
	"github.com/cometbft/cometbft/libs/protoio"
	oracleproto "github.com/cometbft/cometbft/proto/tendermint/oracle"
)

func OracleVoteSignBytes(chainID string, vote *oracleproto.GossipedVotes) []byte {
	pb := CanonicalizeOracleVote(chainID, vote)
	bz, err := protoio.MarshalDelimited(&pb)
	if err != nil {
		panic(err)
	}

	return bz
}

func CanonicalizeOracleVote(chainID string, vote *oracleproto.GossipedVotes) oracleproto.CanonicalGossipedVotes {
	return oracleproto.CanonicalGossipedVotes{
		PubKey:          vote.PubKey,
		Votes:           vote.Votes,
		SignedTimestamp: vote.SignedTimestamp,
		ChainId:         chainID,
	}
}
