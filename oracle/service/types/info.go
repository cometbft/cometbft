package types

import (
	"github.com/cometbft/cometbft/config"
	"github.com/cometbft/cometbft/crypto"
	cmtsync "github.com/cometbft/cometbft/libs/sync"
	oracleproto "github.com/cometbft/cometbft/proto/tendermint/oracle"
	"github.com/cometbft/cometbft/types"
)

// App struct for app
type OracleInfo struct {
	Config             *config.OracleConfig
	UnsignedVoteBuffer *UnsignedVoteBuffer
	GossipVoteBuffer   *GossipVoteBuffer
	SignVotesChan      chan *oracleproto.Vote
	PubKey             crypto.PubKey
	PrivValidator      types.PrivValidator
	StopChannel        chan int
}

type UnsignedVotes struct {
	Timestamp uint64
	Votes     []*oracleproto.Vote
}

type GossipVoteBuffer struct {
	Buffer    map[string]*oracleproto.GossipVote
	UpdateMtx cmtsync.RWMutex
}

type UnsignedVoteBuffer struct {
	Buffer    []*UnsignedVotes // deque of UnsignedVote obj
	UpdateMtx cmtsync.RWMutex
}
