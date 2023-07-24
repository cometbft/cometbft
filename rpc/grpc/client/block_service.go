package client

import (
	"context"

	blocksvc "github.com/cometbft/cometbft/proto/tendermint/services/block/v1"
	"github.com/cometbft/cometbft/types"
	"github.com/cosmos/gogoproto/grpc"
)

// Block data returned by the CometBFT BlockService gRPC API.
type Block struct {
	BlockID types.BlockID `json:"block_id"`
	Block   *types.Block  `json:"block"`
}

// BlockServiceClient provides block information
type BlockServiceClient interface {
	GetBlockByHeight(ctx context.Context, height int64) (*Block, error)
	// GetLatestHeight provides sends the latest committed block height to the given output
	// channel as blocks are committed.
	GetLatestHeight(ctx context.Context, out chan<- int64) error
}

type blockServiceClient struct {
	client blocksvc.BlockServiceClient
}

func newBlockServiceClient(conn grpc.ClientConn) BlockServiceClient {
	return &blockServiceClient{
		client: blocksvc.NewBlockServiceClient(conn),
	}
}

// GetBlockByHeight implements BlockServiceClient GetBlockByHeight
func (c *blockServiceClient) GetBlockByHeight(ctx context.Context, height int64) (*Block, error) {
	req := blocksvc.GetByHeightRequest{
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

	response := Block{
		BlockID: *blockID,
		Block:   block,
	}
	return &response, nil
}

// GetLatestHeight implements BlockServiceClient GetLatestHeight
func (c *blockServiceClient) GetLatestHeight(ctx context.Context, ch chan<- int64) error {
	req := blocksvc.GetLatestHeightRequest{}

	latestHeight, err := c.client.GetLatestHeight(ctx, &req)
	if err != nil {
		return err
	}

	go func() {
		for {
			response, err := latestHeight.Recv()
			if err != nil {
				break
			}
			ch <- response.Height
		}
	}()
	return nil

}

type disabledBlockServiceClient struct{}

func newDisabledBlockServiceClient() BlockServiceClient {
	return &disabledBlockServiceClient{}
}

// GetBlockByHeight implements BlockServiceClient GetBlockByHeight - disabled client
func (*disabledBlockServiceClient) GetBlockByHeight(context.Context, int64) (*Block, error) {
	panic("block service client is disabled")
}

// GetLatestHeight implements BlockServiceClient GetLatestHeight - disabled client
func (*disabledBlockServiceClient) GetLatestHeight(context.Context, chan<- int64) error {
	panic("block service client is disabled")
}
