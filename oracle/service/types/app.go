package types

import (
	// carbonwalletgo "github.com/Switcheo/carbon-wallet-go"
	"github.com/cometbft/cometbft/redis"
	"google.golang.org/grpc"
)

// App struct for app
type App struct {
	Oracles    []Oracle
	AdapterMap map[string]Adapter
	Redis      redis.Service
	// Wallet     carbonwalletgo.Wallet
	Config     Config
	GrpcClient *grpc.ClientConn
}
