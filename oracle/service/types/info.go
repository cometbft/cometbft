package types

import (
	// carbonwalletgo "github.com/Switcheo/carbon-wallet-go"
	"github.com/cometbft/cometbft/redis"
	"google.golang.org/grpc"
)

// App struct for app
type OracleInfo struct {
	Oracles    []Oracle
	AdapterMap map[string]Adapter
	Redis      redis.Service
	Config     Config
	GrpcClient *grpc.ClientConn
	Votes      map[uint64]map[string][]Vote // timestamp : oracleId : []vote
}
