package e2e_test

import (
	"context"
	"testing"
	"time"

	e2e "github.com/cometbft/cometbft/test/e2e/pkg"
	"github.com/cometbft/cometbft/version"
	"github.com/stretchr/testify/require"
)

func TestGRPC_Version(t *testing.T) {
	testNode(t, func(t *testing.T, node e2e.Node) {
		if node.Mode != e2e.ModeFull && node.Mode != e2e.ModeValidator {
			return
		}

		ctx, ctxCancel := context.WithTimeout(context.Background(), time.Minute)
		defer ctxCancel()
		client, err := node.GRPCClient(ctx)
		require.NoError(t, err)

		res, err := client.GetVersion(ctx)
		require.NoError(t, err)

		require.Equal(t, version.TMCoreSemVer, res.Node)
		require.Equal(t, version.ABCIVersion, res.ABCI)
		require.Equal(t, version.P2PProtocol, res.P2P)
		require.Equal(t, version.BlockProtocol, res.Block)
	})
}

func TestGRPC_Block_GetByHeight(t *testing.T) {
	testNode(t, func(t *testing.T, node e2e.Node) {
		if node.Mode != e2e.ModeFull && node.Mode != e2e.ModeValidator {
			return
		}

		blocks := fetchBlockChain(t)

		client, err := node.Client()
		require.NoError(t, err)
		status, err := client.Status(ctx)
		require.NoError(t, err)

		first := status.SyncInfo.EarliestBlockHeight
		last := status.SyncInfo.LatestBlockHeight
		if node.RetainBlocks > 0 {
			first++ // avoid race conditions with block pruning
		}

		ctx, ctxCancel := context.WithTimeout(context.Background(), time.Minute)
		defer ctxCancel()
		gRPCClient, err := node.GRPCClient(ctx)
		require.NoError(t, err)

		for _, block := range blocks {
			if block.Header.Height < first {
				continue
			}
			if block.Header.Height > last {
				break
			}

			// Get first block
			firstBlock, err := gRPCClient.GetBlockByHeight(ctx, first)

			// First block tests
			require.NoError(t, err)
			require.NotNil(t, firstBlock.BlockID)
			require.Equal(t, firstBlock.Block.Height, first)

			// Get last block
			lastBlock, err := gRPCClient.GetBlockByHeight(ctx, last)

			// Last block tests
			require.NoError(t, err)
			require.NotNil(t, lastBlock.BlockID)
			require.Equal(t, lastBlock.Block.Height, last)
		}
	})
}

func TestGRPC_Block_GetLatestHeight(t *testing.T) {

	testnet := loadTestnet(t)
	node := testnet.RandomNode()
	if node.Mode != e2e.ModeFull && node.Mode != e2e.ModeValidator {
		return
	}

	client, err := node.Client()
	require.NoError(t, err)
	status, err := client.Status(ctx)
	require.NoError(t, err)

	gCtx, cancel := context.WithTimeout(context.Background(), time.Minute)
	gRPCClient, err := node.GRPCClient(ctx)
	require.NoError(t, err)

	newHeightChannel, errorChannel := gRPCClient.GetLatestHeight(gCtx)

	count := 0
	for {
		select {
		case <-gCtx.Done():
			require.NoError(t, gCtx.Err())
		case height := <-newHeightChannel:
			require.GreaterOrEqual(t, height, status.SyncInfo.EarliestBlockHeight)
			count++
		case err := <-errorChannel:
			require.Error(t, err)
		}
		if count == 10 {
			cancel()
			return
		}
		if gCtx.Err() != nil {
			require.Error(t, gCtx.Err())
		}
	}
}
