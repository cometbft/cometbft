package client

import (
	"context"
	v1 "github.com/cometbft/cometbft/proto/tendermint/services/block/v1"
	"github.com/cosmos/gogoproto/grpc"
)

// BlockServiceClient provides block information
type BlockServiceClient interface {
	GetBlock(ctx context.Context, request v1.GetBlockRequest) (*v1.GetBlockResponse, error)
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
func (c *blockServiceClient) GetBlock(ctx context.Context, request v1.GetBlockRequest) (*v1.GetBlockResponse, error) {
	res, err := c.client.GetBlock(ctx, &v1.GetBlockRequest{Height: request.Height})
	if err != nil {
		return nil, err
	}
	return res, nil
}

type disabledBlockServiceClient struct{}

func newDisabledBlockServiceClient() BlockServiceClient {
	return &disabledBlockServiceClient{}
}

// GetBlock implements BlockServiceClient
func (*disabledBlockServiceClient) GetBlock(context.Context, v1.GetBlockRequest) (*v1.GetBlockResponse, error) {
	panic("block service client is disabled")
}
