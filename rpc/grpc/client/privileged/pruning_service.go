package privileged

import (
	"context"

	"github.com/cosmos/gogoproto/grpc"

	v1 "github.com/cometbft/cometbft/proto/tendermint/services/pruning/v1"
)

// RetainHeights provides information on which block height limits have been
// set for block information to be retained by the ABCI application and the
// pruning service.
type RetainHeights struct {
	App            uint64
	PruningService uint64
}

type PruningServiceClient interface {
	SetBlockRetainHeight(ctx context.Context, height uint64) error
	GetBlockRetainHeight(ctx context.Context) (RetainHeights, error)
	SetBlockResultsRetainHeight(ctx context.Context, height uint64) error
	GetBlockResultsRetainHeight(ctx context.Context) (uint64, error)
}

type pruningServiceClient struct {
	inner v1.PruningServiceClient
}

func newPruningServiceClient(conn grpc.ClientConn) PruningServiceClient {
	return &pruningServiceClient{
		inner: v1.NewPruningServiceClient(conn),
	}
}

// SetBlockRetainHeight implements PruningServiceClient.
func (c *pruningServiceClient) SetBlockRetainHeight(ctx context.Context, height uint64) error {
	_, err := c.inner.SetBlockRetainHeight(ctx, &v1.SetBlockRetainHeightRequest{
		Height: height,
	})
	return err
}

// GetBlockRetainHeight implements PruningServiceClient.
func (c *pruningServiceClient) GetBlockRetainHeight(ctx context.Context) (RetainHeights, error) {
	res, err := c.inner.GetBlockRetainHeight(ctx, &v1.GetBlockRetainHeightRequest{})
	if err != nil {
		return RetainHeights{}, err
	}
	return RetainHeights{
		App:            res.AppRetainHeight,
		PruningService: res.PruningServiceRetainHeight,
	}, nil
}

// SetBlockResultsRetainHeight implements PruningServiceClient.
func (c *pruningServiceClient) SetBlockResultsRetainHeight(ctx context.Context, height uint64) error {
	_, err := c.inner.SetBlockResultsRetainHeight(ctx, &v1.SetBlockResultsRetainHeightRequest{
		Height: height,
	})
	return err
}

// GetBlockResultsRetainHeight implements PruningServiceClient.
func (c *pruningServiceClient) GetBlockResultsRetainHeight(ctx context.Context) (uint64, error) {
	res, err := c.inner.GetBlockResultsRetainHeight(ctx, &v1.GetBlockResultsRetainHeightRequest{})
	if err != nil {
		return 0, err
	}
	return res.PruningServiceRetainHeight, nil
}

type disabledPruningServiceClient struct{}

func newDisabledPruningServiceClient() PruningServiceClient {
	return &disabledPruningServiceClient{}
}

// SetBlockRetainHeight implements PruningServiceClient.
func (*disabledPruningServiceClient) SetBlockRetainHeight(ctx context.Context, height uint64) error {
	panic("pruning service client is disabled")
}

// GetBlockRetainHeight implements PruningServiceClient.
func (*disabledPruningServiceClient) GetBlockRetainHeight(ctx context.Context) (RetainHeights, error) {
	panic("pruning service client is disabled")
}

// SetBlockResultsRetainHeight implements PruningServiceClient.
func (*disabledPruningServiceClient) SetBlockResultsRetainHeight(ctx context.Context, height uint64) error {
	panic("pruning service client is disabled")
}

// GetBlockResultsRetainHeight implements PruningServiceClient.
func (*disabledPruningServiceClient) GetBlockResultsRetainHeight(ctx context.Context) (uint64, error) {
	panic("pruning service client is disabled")
}
