package core

import (
	ctypes "github.com/tendermint/tendermint/rpc/core/types"
	rpctypes "github.com/tendermint/tendermint/rpc/jsonrpc/types"
)

// UnsafeFlushMempool removes all transactions from the mempool.
<<<<<<< HEAD
func UnsafeFlushMempool(ctx *rpctypes.Context) (*ctypes.ResultUnsafeFlushMempool, error) {
=======
func (env *Environment) UnsafeFlushMempool(*rpctypes.Context) (*ctypes.ResultUnsafeFlushMempool, error) {
>>>>>>> 111d252d7 (Fix lints (#625))
	env.Mempool.Flush()
	return &ctypes.ResultUnsafeFlushMempool{}, nil
}
