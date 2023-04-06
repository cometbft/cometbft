package coregrpc

import (
	"context"

	"github.com/cosmos/gogoproto/grpc"

	abci "github.com/cometbft/cometbft/abci/types"
	v1 "github.com/cometbft/cometbft/api/cometbft/rpc/grpc/v1"
	v3 "github.com/cometbft/cometbft/api/cometbft/rpc/grpc/v3"
	core "github.com/cometbft/cometbft/rpc/core"
	rpctypes "github.com/cometbft/cometbft/rpc/jsonrpc/types"
)

type BroadcastAPIClient = v3.BroadcastAPIClient
type BroadcastAPIServer = v3.BroadcastAPIServer
type RequestBroadcastTx = v1.RequestBroadcastTx
type RequestPing = v1.RequestPing
type ResponseBroadcastTx = v3.ResponseBroadcastTx
type ResponsePing = v1.ResponsePing

func NewBroadcastAPIClient(cc grpc.ClientConn) BroadcastAPIClient {
	return v3.NewBroadcastAPIClient(cc)
}

func RegisterBroadcastAPIServer(s grpc.Server, srv BroadcastAPIServer) {
	v3.RegisterBroadcastAPIServer(s, srv)
}

type broadcastAPI struct {
	env *core.Environment
}

func (bapi *broadcastAPI) Ping(ctx context.Context, req *RequestPing) (*ResponsePing, error) {
	// kvstore so we can check if the server is up
	return &ResponsePing{}, nil
}

func (bapi *broadcastAPI) BroadcastTx(ctx context.Context, req *RequestBroadcastTx) (*ResponseBroadcastTx, error) {
	// NOTE: there's no way to get client's remote address
	// see https://stackoverflow.com/questions/33684570/session-and-remote-ip-address-in-grpc-go
	res, err := bapi.env.BroadcastTxCommit(&rpctypes.Context{}, req.Tx)
	if err != nil {
		return nil, err
	}

	return &ResponseBroadcastTx{
		CheckTx: &abci.ResponseCheckTx{
			Code: res.CheckTx.Code,
			Data: res.CheckTx.Data,
			Log:  res.CheckTx.Log,
		},
		TxResult: &abci.ExecTxResult{
			Code: res.TxResult.Code,
			Data: res.TxResult.Data,
			Log:  res.TxResult.Log,
		},
	}, nil
}
