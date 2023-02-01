package core

import (
	ctypes "github.com/cometbft/cometbft/rpc/core/types"
	rpctypes "github.com/cometbft/cometbft/rpc/jsonrpc/types"
)

// Health gets node health. Returns empty result (200 OK) on success, no
// response - in case of an error.
<<<<<<< HEAD
// More: https://docs.tendermint.com/v0.37/rpc/#/Info/health
func Health(ctx *rpctypes.Context) (*ctypes.ResultHealth, error) {
=======
// More: https://docs.cometbft.com/main/rpc/#/Info/health
func (env *Environment) Health(ctx *rpctypes.Context) (*ctypes.ResultHealth, error) {
>>>>>>> 1cb55d49b (Rename Tendermint to CometBFT: further actions (#224))
	return &ctypes.ResultHealth{}, nil
}
