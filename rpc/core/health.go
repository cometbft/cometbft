package core

import (
	ctypes "github.com/cometbft/cometbft/rpc/core/types"
	rpctypes "github.com/cometbft/cometbft/rpc/jsonrpc/types"
)

// Health gets node health. Returns empty result (200 OK) on success, no
// response - in case of an error.
<<<<<<< HEAD
// More: https://docs.cometbft.com/v0.38.x/rpc/#/Info/health
func (env *Environment) Health(ctx *rpctypes.Context) (*ctypes.ResultHealth, error) {
=======
// More: https://docs.cometbft.com/main/rpc/#/Info/health
func (env *Environment) Health(*rpctypes.Context) (*ctypes.ResultHealth, error) {
>>>>>>> 111d252d7 (Fix lints (#625))
	return &ctypes.ResultHealth{}, nil
}
