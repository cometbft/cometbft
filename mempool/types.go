package mempool

import (
	memprotos "github.com/cometbft/cometbft/api/cometbft/mempool/v1beta1"
	"github.com/cometbft/cometbft/p2p"
)

var (
	_ p2p.Wrapper   = &memprotos.Txs{}
	_ p2p.Unwrapper = &memprotos.Message{}
)
