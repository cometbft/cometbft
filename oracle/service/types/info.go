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
	Oracles            []Oracle
	AdapterMap         map[string]Adapter
	Redis              redis.Service
	Config             Config
	GrpcClient         *grpc.ClientConn
	UnsignedVoteBuffer *UnsignedVoteBuffer
	GossipVoteBuffer   *GossipVoteBuffer
	SignVotesChan      chan *oracleproto.Vote
	PubKey             crypto.PubKey
	PrivValidator      types.PrivValidator
	MsgFlushInterval   time.Duration
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
