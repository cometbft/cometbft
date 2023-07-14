package client

import (
	"context"

	blocksvc "github.com/cometbft/cometbft/proto/tendermint/services/block/v1"
	"github.com/cometbft/cometbft/types"
	"github.com/cosmos/gogoproto/grpc"
)

// ResultBlock Single block (with meta)
type ResultBlock struct {
	BlockID types.BlockID `json:"block_id"`
	Block   *types.Block  `json:"block"`
}

// BlockServiceClient provides block information
type BlockServiceClient interface {
	GetBlockByHeight(ctx context.Context, height int64) (*ResultBlock, error)
}

type blockServiceClient struct {
	client blocksvc.BlockServiceClient
}

func newBlockServiceClient(conn grpc.ClientConn) BlockServiceClient {
	return &blockServiceClient{
		client: blocksvc.NewBlockServiceClient(conn),
	}
}

// GetBlockByHeight implements BlockServiceClient
func (c *blockServiceClient) GetBlockByHeight(ctx context.Context, height int64) (*ResultBlock, error) {
	req := blocksvc.GetBlockByHeightRequest{
		Height: height,
	}
	res, err := c.client.GetByHeight(ctx, &req)
	if err != nil {
		return nil, err
	}

	// convert Block from proto to core type
	block, err := types.BlockFromProto(res.Block)
	if err != nil {
		return nil, err
	}

	// convert BlockID from proto to core type
	blockID, err := types.BlockIDFromProto(res.BlockId)
	if err != nil {
		return nil, err
	}

	response := ResultBlock{
		BlockID: *blockID,
		Block:   block,
	}
	return &response, nil
}

type disabledBlockServiceClient struct{}

func newDisabledBlockServiceClient() BlockServiceClient {
	return &disabledBlockServiceClient{}
}

// GetBlockByHeight implements BlockServiceClient
func (*disabledBlockServiceClient) GetBlockByHeight(context.Context, int64) (*ResultBlock, error) {
	panic("block service client is disabled")
}
