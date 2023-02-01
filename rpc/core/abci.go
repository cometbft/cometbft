package core

import (
	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/cometbft/cometbft/libs/bytes"
	"github.com/cometbft/cometbft/proxy"
	ctypes "github.com/cometbft/cometbft/rpc/core/types"
	rpctypes "github.com/cometbft/cometbft/rpc/jsonrpc/types"
)

// ABCIQuery queries the application for some information.
<<<<<<< HEAD
// More: https://docs.tendermint.com/v0.34/rpc/#/ABCI/abci_query
func ABCIQuery(
=======
// More: https://docs.cometbft.com/main/rpc/#/ABCI/abci_query
func (env *Environment) ABCIQuery(
>>>>>>> 1cb55d49b (Rename Tendermint to CometBFT: further actions (#224))
	ctx *rpctypes.Context,
	path string,
	data bytes.HexBytes,
	height int64,
	prove bool,
) (*ctypes.ResultABCIQuery, error) {
	resQuery, err := env.ProxyAppQuery.QuerySync(abci.RequestQuery{
		Path:   path,
		Data:   data,
		Height: height,
		Prove:  prove,
	})
	if err != nil {
		return nil, err
	}

	return &ctypes.ResultABCIQuery{Response: *resQuery}, nil
}

// ABCIInfo gets some info about the application.
<<<<<<< HEAD
// More: https://docs.tendermint.com/v0.34/rpc/#/ABCI/abci_info
func ABCIInfo(ctx *rpctypes.Context) (*ctypes.ResultABCIInfo, error) {
=======
// More: https://docs.cometbft.com/main/rpc/#/ABCI/abci_info
func (env *Environment) ABCIInfo(ctx *rpctypes.Context) (*ctypes.ResultABCIInfo, error) {
>>>>>>> 1cb55d49b (Rename Tendermint to CometBFT: further actions (#224))
	resInfo, err := env.ProxyAppQuery.InfoSync(proxy.RequestInfo)
	if err != nil {
		return nil, err
	}

	return &ctypes.ResultABCIInfo{Response: *resInfo}, nil
}
