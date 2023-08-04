package client

import (
	"context"
	"fmt"

	blocksvc "github.com/cometbft/cometbft/proto/tendermint/services/block/v1"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/cometbft/cometbft/types"
	"github.com/cosmos/gogoproto/grpc"
)

// Block data returned by the CometBFT BlockService gRPC API.
type Block struct {
	BlockID *types.BlockID `json:"block_id"`
	Block   *types.Block   `json:"block"`
}

func blockFromProto(pblockID *cmtproto.BlockID, pblock *cmtproto.Block) (*Block, error) {
	blockID, err := types.BlockIDFromProto(pblockID)
	if err != nil {
		return nil, err
	}

	block, err := types.BlockFromProto(pblock)
	if err != nil {
		return nil, err
	}

	return &Block{
		BlockID: blockID,
		Block:   block,
	}, nil
}

// LatestHeightResult type used in GetLatestResult and send to the client
// via a channel
type LatestHeightResult struct {
	Height int64
	Error  error
}

// BlockServiceClient provides block information
type BlockServiceClient interface {
	// GetBlockByHeight attempts to retrieve the block associated with the
	// given height.
	GetBlockByHeight(ctx context.Context, height int64) (*Block, error)

	// GetLatestBlock attempts to retrieve the latest committed block.
	GetLatestBlock(ctx context.Context) (*Block, error)

	// GetLatestHeight provides sends the latest committed block height to the
	// given output channel as blocks are committed.
	//
	// Only returns an error if request initiation fails, otherwise errors are
	// returned via the supplied channel.
	GetLatestHeight(ctx context.Context, resultCh chan<- LatestHeightResult) error
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
	res, err := c.client.GetByHeight(ctx, &blocksvc.GetByHeightRequest{
		Height: height,
	})
	if err != nil {
		return nil, err
	}

	return blockFromProto(res.BlockId, res.Block)
}

// GetLatestBlock implements BlockServiceClient.
func (c *blockServiceClient) GetLatestBlock(ctx context.Context) (*Block, error) {
	res, err := c.client.GetLatest(ctx, &blocksvc.GetLatestRequest{})
	if err != nil {
		return nil, err
	}

	return blockFromProto(res.BlockId, res.Block)
}

// GetLatestHeight implements BlockServiceClient GetLatestHeight
// This method provides an out channel (int64) that streams the latest height.
// The out channel might return non-contiguous heights if the channel becomes full,
func (c *blockServiceClient) GetLatestHeight(ctx context.Context, resultCh chan<- LatestHeightResult) error {
	req := blocksvc.GetLatestHeightRequest{}

	latestHeightClient, err := c.client.GetLatestHeight(ctx, &req)
	if err != nil {
		return fmt.Errorf("error getting a stream for the latest height: %w", err)
	}

	go func(client blocksvc.BlockService_GetLatestHeightClient) {
		for {
			response, err := client.Recv()
			if err != nil {
				resultCh <- LatestHeightResult{
					Height: 0,
					Error:  fmt.Errorf("error receiving the latest height from a stream: %w", err),
				}
				break
			}
			res := LatestHeightResult{Height: response.Height}
			select {
			case resultCh <- res:
			default:
				// Skip sending this result because the channel is full - the
				// client will get the next one once the channel opens up again
			}

		}
	}(latestHeightClient)
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

// GetLatestBlock implements BlockServiceClient.
func (*disabledBlockServiceClient) GetLatestBlock(context.Context) (*Block, error) {
	panic("block service client is disabled")
}

// GetLatestHeight implements BlockServiceClient GetLatestHeight - disabled client
func (*disabledBlockServiceClient) GetLatestHeight(context.Context, chan<- LatestHeightResult) error {
	panic("block service client is disabled")
}
