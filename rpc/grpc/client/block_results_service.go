package client

import (
	"context"
	"fmt"

	abci "github.com/cometbft/cometbft/abci/types"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"

	"github.com/cosmos/gogoproto/grpc"

	brs "github.com/cometbft/cometbft/proto/tendermint/services/block_results/v1"
)

type ResultBlockResults struct {
	Height                int64                     `json:"height"`
	TxsResults            []*abci.ExecTxResult      `json:"txs_results"`
	FinalizeBlockEvents   []*abci.Event             `json:"finalize_block_events"`
	ValidatorUpdates      []*abci.ValidatorUpdate   `json:"validator_updates"`
	ConsensusParamUpdates *cmtproto.ConsensusParams `json:"consensus_param_updates"`
	AppHash               []byte                    `json:"app_hash"`
}

// BlockResultsServiceClient provides the block results of a given height (or latest if none provided).
type BlockResultsServiceClient interface {
	GetBlockResults(ctx context.Context, req brs.GetBlockResultsRequest) (*ResultBlockResults, error)
}

type blockResultServiceClient struct {
	client brs.BlockResultsServiceClient
}

func (b blockResultServiceClient) GetBlockResults(ctx context.Context, req brs.GetBlockResultsRequest) (*ResultBlockResults, error) {
	res, err := b.client.GetBlockResults(ctx, &brs.GetBlockResultsRequest{Height: req.Height})
	if err != nil {
		return nil, fmt.Errorf("error fetching BlockResults :: %s", err.Error())
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
		client: brs.NewBlockResultsServiceClient(conn),
	}
}

type disabledBlockResultsServiceClient struct{}

func newDisabledBlockResultsServiceClient() BlockResultsServiceClient {
	return &disabledBlockResultsServiceClient{}
}

// GetBLockResults implements BlockResultsServiceClient
func (*disabledBlockResultsServiceClient) GetBlockResults(_ context.Context, _ brs.GetBlockResultsRequest) (*ResultBlockResults, error) {
	panic("block results service client is disabled")
}
