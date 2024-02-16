package types

import (
	"time"

	"github.com/cometbft/cometbft/crypto"
	cmtsync "github.com/cometbft/cometbft/libs/sync"
	oracleproto "github.com/cometbft/cometbft/proto/tendermint/oracle"
	"github.com/cometbft/cometbft/redis"
	"github.com/cometbft/cometbft/types"
	"google.golang.org/grpc"
)

// App struct for app
type OracleInfo struct {
	Oracles          []Oracle
	AdapterMap       map[string]Adapter
	Redis            redis.Service
	Config           Config
	GrpcClient       *grpc.ClientConn
	VoteDataBuffer   *VoteDataBuffer
	GossipVoteBuffer *GossipVoteBuffer
	SignVotesChan    chan *oracleproto.Vote
	PubKey           crypto.PubKey
	PrivValidator    types.PrivValidator
	MsgFlushInterval time.Duration
	StopChannel      chan int
}

type GossipVoteBuffer struct {
	Buffer    map[uint64]*oracleproto.GossipVote
	UpdateMtx cmtsync.RWMutex
}

type VoteDataBuffer struct {
	Buffer    map[uint64]map[string][]*oracleproto.Vote
	UpdateMtx cmtsync.RWMutex
}
