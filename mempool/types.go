package mempool

import (
	memprotos "github.com/cometbft/cometbft/api/cometbft/mempool/v2"
	"github.com/cometbft/cometbft/v2/types"
)

var (
	_ types.Wrapper   = &memprotos.Txs{}
	_ types.Wrapper   = &memprotos.HaveTx{}
	_ types.Wrapper   = &memprotos.ResetRoute{}
	_ types.Unwrapper = &memprotos.Message{}
)
