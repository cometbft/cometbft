package client

import (
	"context"
	"fmt"

	blocksvc "github.com/cometbft/cometbft/proto/tendermint/services/block/v1"
	"github.com/cometbft/cometbft/types"
	"github.com/cosmos/gogoproto/grpc"
)

// Block data returned by the CometBFT BlockService gRPC API.
type Block struct {
	BlockID types.BlockID `json:"block_id"`
	Block   *types.Block  `json:"block"`
}

// LatestHeightResult type used in GetLatestResult and send to the client
// via a channel
type LatestHeightResult struct {
	Height int64
	Error  error
}

// BlockServiceClient provides block information
type BlockServiceClient interface {
	GetBlockByHeight(ctx context.Context, height int64) (*Block, error)
	// GetLatestHeight provides sends the latest committed block height to the given output
	// channel as blocks are committed.
	GetLatestHeight(ctx context.Context, resultCh chan<- LatestHeightResult)
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
// This method provides an out channel (int64) that streams the latest height.
// The out channel might return non-contiguous heights if the channel becomes full,
func (c *blockServiceClient) GetLatestHeight(ctx context.Context, resultCh chan<- LatestHeightResult) {
	req := blocksvc.GetLatestHeightRequest{}

	latestHeightClient, err := c.client.GetLatestHeight(ctx, &req)
	if err != nil {
		resultCh <- LatestHeightResult{
			Height: 0,
			Error:  fmt.Errorf("error getting a stream for the latest height"),
		}
	}

	go func(client blocksvc.BlockService_GetLatestHeightClient) {
		for {
			response, err := client.Recv()
			if err != nil {
				resultCh <- LatestHeightResult{
					Height: 0,
					Error:  fmt.Errorf("error receiving the latest height from a stream"),
				}
				break
			}
			select {
			case resultCh <- LatestHeightResult{
				Height: response.Height,
				Error:  fmt.Errorf("error receiving the latest height from a stream"),
			}:
			default:
			}

		}
	}(latestHeightClient)
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
func (*disabledBlockServiceClient) GetLatestHeight(context.Context, chan<- LatestHeightResult) {
	panic("block service client is disabled")
}
