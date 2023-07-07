package client

import (
	"context"
	cmtversion "github.com/cometbft/cometbft/proto/tendermint/version"
	"time"

	v1 "github.com/cometbft/cometbft/proto/tendermint/services/block/v1"
	"github.com/cometbft/cometbft/types"
	"github.com/cosmos/gogoproto/grpc"
)

// BlockServiceClient provides block information
type BlockServiceClient interface {
	GetBlock(ctx context.Context, height int64) (*types.Block, error)
}

type blockServiceClient struct {
	client v1.BlockServiceClient
}

func newBlockServiceClient(conn grpc.ClientConn) BlockServiceClient {
	return &blockServiceClient{
		client: v1.NewBlockServiceClient(conn),
	}
}

// GetBlock implements BlockServiceClient
func (c *blockServiceClient) GetBlock(ctx context.Context, height int64) (*types.Block, error) {
	res, err := c.client.GetBlock(ctx, &v1.GetBlockRequest{Height: height})
	if err != nil {
		return nil, err
	}
	return &types.Block{
		Header: types.Header{
			Version:            cmtversion.Consensus{},
			ChainID:            "",
			Height:             res.Height,
			Time:               time.Time{},
			LastBlockID:        types.BlockID{},
			LastCommitHash:     nil,
			DataHash:           nil,
			ValidatorsHash:     nil,
			NextValidatorsHash: nil,
			ConsensusHash:      nil,
			AppHash:            nil,
			LastResultsHash:    nil,
			EvidenceHash:       nil,
			ProposerAddress:    nil,
		},
		Data:       types.Data{},
		Evidence:   types.EvidenceData{},
		LastCommit: nil,
	}, nil
}

type disabledBlockServiceClient struct{}

func newDisabledBlockServiceClient() BlockServiceClient {
	return &disabledBlockServiceClient{}
}

// GetBlock implements BlockServiceClient
func (*disabledBlockServiceClient) GetBlock(context.Context, int64) (*types.Block, error) {
	panic("block service client is disabled")
}
