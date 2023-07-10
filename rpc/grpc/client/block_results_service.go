package client

import (
	"context"

	abci "github.com/cometbft/cometbft/abci/types"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"

	"github.com/cosmos/gogoproto/grpc"

	v1 "github.com/cometbft/cometbft/proto/tendermint/services/block_results/v1"
)

type ResultBlockResults struct {
	Height                int64                     `json:"height"`
	TxsResults            []*abci.ExecTxResult      `json:"txs_results"`
	FinalizeBlockEvents   []abci.Event              `json:"finalize_block_events"`
	ValidatorUpdates      []abci.ValidatorUpdate    `json:"validator_updates"`
	ConsensusParamUpdates *cmtproto.ConsensusParams `json:"consensus_param_updates"`
	AppHash               []byte                    `json:"app_hash"`
}

// BlockResultsServiceClient provides the block results of a given height (or latest if none provided).
type BlockResultsServiceClient interface {
	GetBlockResults(ctx context.Context, req v1.GetBlockResultsRequest) (*ResultBlockResults, error)
}

type blockResultServiceClient struct {
	client v1.BlockResultsServiceClient
}

func (b blockResultServiceClient) GetBlockResults(ctx context.Context, req v1.GetBlockResultsRequest) (*ResultBlockResults, error) {
	res, err := b.client.GetBlockResults(ctx, &v1.GetBlockResultsRequest{Height: req.Height})
	if err != nil {
		return &ResultBlockResults{}, err
	}

	return &ResultBlockResults{
		Height:                res.Height,
		TxsResults:            res.TxsResults,
		FinalizeBlockEvents:   res.FinalizeBlockEvents,
		ValidatorUpdates:      res.ValidatorUpdates,
		ConsensusParamUpdates: res.ConsensusParamUpdates,
		AppHash:               res.AppHash,
	}, nil
}

func newBlockResultsServiceClient(conn grpc.ClientConn) BlockResultsServiceClient {
	return &blockResultServiceClient{
		client: v1.NewBlockResultsServiceClient(conn),
	}
}

type disabledBlockResultsServiceClient struct{}

func newDisabledBlockResultsServiceClient() BlockResultsServiceClient {
	return &disabledBlockResultsServiceClient{}
}

// GetBLockResults implements BlockResultsServiceClient
func (*disabledBlockResultsServiceClient) GetBlockResults(_ context.Context, _ v1.GetBlockResultsRequest) (*ResultBlockResults, error) {
	panic("block results service client is disabled")
}
