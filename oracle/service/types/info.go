package types

import (
	"github.com/cometbft/cometbft/config"
	"github.com/cometbft/cometbft/crypto"
	cmtsync "github.com/cometbft/cometbft/libs/sync"
	oracleproto "github.com/cometbft/cometbft/proto/tendermint/oracle"
	"github.com/cometbft/cometbft/proxy"
	"github.com/cometbft/cometbft/types"
)

// App struct for app
type OracleInfo struct {
	Config           *config.OracleConfig
	GossipVoteBuffer *GossipVoteBuffer
	SignVotesChan    chan *oracleproto.Vote
	PubKey           crypto.PubKey
	PrivValidator    types.PrivValidator
	StopChannel      chan int
	ProxyApp         proxy.AppConnConsensus
	BlockTimestamps  []int64
}
type GossipVoteBuffer struct {
	Buffer    map[string]*oracleproto.GossipVote
	UpdateMtx cmtsync.RWMutex
}
