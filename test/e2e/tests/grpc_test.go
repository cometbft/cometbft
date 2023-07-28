package e2e_test

import (
	"context"
	"testing"
	"time"

	v1 "github.com/cometbft/cometbft/proto/tendermint/services/block_results/v1"
	client2 "github.com/cometbft/cometbft/rpc/grpc/client"

	"github.com/stretchr/testify/require"

	e2e "github.com/cometbft/cometbft/test/e2e/pkg"
	"github.com/cometbft/cometbft/version"
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

	var node *e2e.Node
	testnet := loadTestnet(t)
	// this assumes at least one node is valid for testing
	// not a light or seed node
	for _, n := range testnet.ArchiveNodes() {
		if !n.Stateless() {
			node = n
			break
		}
	}
	require.NotNil(t, node)

	client, err := node.Client()
	require.NoError(t, err)
	status, err := client.Status(ctx)
	require.NoError(t, err)

	gCtx, cancel := context.WithTimeout(context.Background(), time.Minute)
	gRPCClient, err := node.GRPCClient(ctx)
	require.NoError(t, err)

	resultCh := make(chan client2.LatestHeightResult)
	gRPCClient.GetLatestHeight(gCtx, resultCh)

	count := 0
	for {
		select {
		case <-gCtx.Done():
			require.NoError(t, gCtx.Err())
		case result := <-resultCh:
			if err != nil {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.GreaterOrEqual(t, result.Height, status.SyncInfo.EarliestBlockHeight)
				count++
			}
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

func TestGRPC_GetBlockResults(t *testing.T) {
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
			first++
		}

		ctx, ctxCancel := context.WithTimeout(context.Background(), time.Minute)
		defer ctxCancel()
		gRPCClient, err := node.GRPCClient(ctx)
		require.NoError(t, err)

		for _, block := range blocks {
			successCases := []struct {
				expectedHeight int64
				request        v1.GetBlockResultsRequest
			}{
				{last, v1.GetBlockResultsRequest{}},
				{first, v1.GetBlockResultsRequest{Height: first}},
			}
			errorCases := []struct {
				request v1.GetBlockResultsRequest
			}{
				{v1.GetBlockResultsRequest{Height: -1}},
				{v1.GetBlockResultsRequest{Height: 10000}},
			}

			if block.Header.Height < first {
				continue
			}
			if block.Header.Height > last {
				break
			}
			for _, tc := range successCases {
				res, err := gRPCClient.GetBlockResults(ctx, tc.request)
				// First block tests
				require.NoError(t, err)
				require.NotNil(t, res)
				if tc.expectedHeight == last {
					require.GreaterOrEqual(t, tc.expectedHeight, res.Height)
				} else {
					require.Equal(t, res.Height, tc.expectedHeight)
				}
			}
			for _, tc := range errorCases {
				_, err = gRPCClient.GetBlockResults(ctx, tc.request)
				require.Error(t, err)
			}
		}
	})
}
