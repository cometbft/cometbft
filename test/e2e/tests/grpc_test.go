package e2e_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	coretypes "github.com/cometbft/cometbft/rpc/core/types"
	"github.com/cometbft/cometbft/rpc/grpc/client/privileged"
	e2e "github.com/cometbft/cometbft/test/e2e/pkg"
	"github.com/cometbft/cometbft/version"
)

func TestGRPC_Version(t *testing.T) {
	testFullNodesOrValidators(t, 0, func(t *testing.T, node e2e.Node) {
		ctx, ctxCancel := context.WithTimeout(context.Background(), time.Minute)
		defer ctxCancel()
		client, err := node.GRPCClient(ctx)
		require.NoError(t, err)
		defer client.Close()

		res, err := client.GetVersion(ctx)
		require.NoError(t, err)

		require.Equal(t, version.CMTSemVer, res.Node)
		require.Equal(t, version.ABCIVersion, res.ABCI)
		require.Equal(t, version.P2PProtocol, res.P2P)
		require.Equal(t, version.BlockProtocol, res.Block)
	})
}

func TestGRPC_Block_GetByHeight(t *testing.T) {
	testFullNodesOrValidators(t, 0, func(t *testing.T, node e2e.Node) {

		client, err := node.Client()
		require.NoError(t, err)
		status, err := client.Status(ctx)
		require.NoError(t, err)

		// We are not testing getting the first block in these
		// tests to prevent race conditions with the pruning mechanism
		// that might make the tests fail. Just testing the last block
		// is enough to validate the fact that we can fetch a block using
		// the gRPC endpoint
		last := status.SyncInfo.LatestBlockHeight

		ctx, ctxCancel := context.WithTimeout(context.Background(), time.Minute)
		defer ctxCancel()
		gRPCClient, err := node.GRPCClient(ctx)
		require.NoError(t, err)
		defer gRPCClient.Close()

		// Get last block and fetch it using the gRPC endpoint
		lastBlock, err := gRPCClient.GetBlockByHeight(ctx, last)

		// Last block tests
		require.NoError(t, err)
		require.NotNil(t, lastBlock.BlockID)
		require.Equal(t, lastBlock.Block.Height, last)

	})
}

func TestGRPC_Block_GetLatest(t *testing.T) {
	testFullNodesOrValidators(t, 1, func(t *testing.T, node e2e.Node) {
		if node.Mode != e2e.ModeFull && node.Mode != e2e.ModeValidator {
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
		defer cancel()

		gclient, err := node.GRPCClient(ctx)
		require.NoError(t, err)
		defer gclient.Close()

		resultCh, err := gclient.GetLatestHeight(ctx)
		require.NoError(t, err)

		select {
		case <-ctx.Done():
			require.Fail(t, "did not expect context to be canceled")
		case result := <-resultCh:
			require.NoError(t, result.Error)
			block, err := gclient.GetLatestBlock(ctx)
			require.NoError(t, err)
			// We can be off by at most one block, depending on how quickly the
			// latest block request was executed.
			require.True(t, result.Height == block.Block.Height || result.Height == block.Block.Height+1)
		}
	})
}

func TestGRPC_Block_GetLatestHeight(t *testing.T) {
	testFullNodesOrValidators(t, 0, func(t *testing.T, node e2e.Node) {
		ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
		defer cancel()

		client, err := node.Client()
		require.NoError(t, err)
		status, err := client.Status(ctx)
		require.NoError(t, err)

		gclient, err := node.GRPCClient(ctx)
		require.NoError(t, err)
		defer gclient.Close()

		resultCh, err := gclient.GetLatestHeight(ctx)
		require.NoError(t, err)

		select {
		case <-ctx.Done():
			require.Fail(t, "did not expect context to be canceled")
		case result := <-resultCh:
			require.NoError(t, result.Error)
			require.True(t, result.Height == status.SyncInfo.LatestBlockHeight || result.Height == status.SyncInfo.LatestBlockHeight+1)
		}
	})
}

func TestGRPC_GetBlockResults(t *testing.T) {
	testFullNodesOrValidators(t, 0, func(t *testing.T, node e2e.Node) {
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
		defer gRPCClient.Close()

		// GetLatestBlockResults
		latestBlockResults, err := gRPCClient.GetLatestBlockResults(ctx)

		require.GreaterOrEqual(t, last, latestBlockResults.Height)
		require.NoError(t, err, "Unexpected error for GetLatestBlockResults")
		require.NotNil(t, latestBlockResults)

		successCases := []struct {
			expectedHeight int64
		}{
			{first},
			{latestBlockResults.Height},
		}
		errorCases := []struct {
			requestHeight int64
		}{
			{first - 2},
			{last + 100000},
		}

		for _, tc := range successCases {
			res, err := gRPCClient.GetBlockResults(ctx, tc.expectedHeight)

			require.NoError(t, err, fmt.Sprintf("Unexpected error for GetBlockResults at expected height: %d", tc.expectedHeight))
			require.NotNil(t, res)
			require.Equal(t, res.Height, tc.expectedHeight)
		}
		for _, tc := range errorCases {
			_, err = gRPCClient.GetBlockResults(ctx, tc.requestHeight)
			require.Error(t, err)
		}
	})
}

func TestGRPC_BlockRetainHeight(t *testing.T) {
	testFullNodesOrValidators(t, 0, func(t *testing.T, node e2e.Node) {
		if !node.EnableCompanionPruning {
			return
		}

		grpcClient, status, cleanup := getGRPCPrivilegedClientForTesting(t, node)
		defer cleanup()

		err := grpcClient.SetBlockRetainHeight(ctx, uint64(status.SyncInfo.LatestBlockHeight-1))
		require.NoError(t, err)

		res, err := grpcClient.GetBlockRetainHeight(ctx)
		require.NoError(t, err)
		require.NotNil(t, res)
		require.Equal(t, res.PruningService, uint64(status.SyncInfo.LatestBlockHeight-1))
	})
}

func TestGRPC_BlockResultsRetainHeight(t *testing.T) {
	testFullNodesOrValidators(t, 0, func(t *testing.T, node e2e.Node) {
		if !node.EnableCompanionPruning {
			return
		}

		grpcClient, status, cleanup := getGRPCPrivilegedClientForTesting(t, node)
		defer cleanup()

		err := grpcClient.SetBlockResultsRetainHeight(ctx, uint64(status.SyncInfo.LatestBlockHeight)-1)
		require.NoError(t, err, "Unexpected error for SetBlockResultsRetainHeight")

		height, err := grpcClient.GetBlockResultsRetainHeight(ctx)
		require.NoError(t, err, "Unexpected error for GetBlockRetainHeight")
		require.Equal(t, height, uint64(status.SyncInfo.LatestBlockHeight)-1)
	})
}

func TestGRPC_TxIndexerRetainHeight(t *testing.T) {
	testFullNodesOrValidators(t, 0, func(t *testing.T, node e2e.Node) {
		if !node.EnableCompanionPruning {
			return
		}

		grpcClient, status, cleanup := getGRPCPrivilegedClientForTesting(t, node)
		defer cleanup()

		err := grpcClient.SetTxIndexerRetainHeight(ctx, uint64(status.SyncInfo.LatestBlockHeight)-1)
		require.NoError(t, err, "Unexpected error for SetTxIndexerRetainHeight")

		height, err := grpcClient.GetTxIndexerRetainHeight(ctx)
		require.NoError(t, err, "Unexpected error for GetTxIndexerRetainHeight")
		require.Equal(t, height, uint64(status.SyncInfo.LatestBlockHeight)-1)
	})
}

func TestGRPC_BlockIndexerRetainHeight(t *testing.T) {
	testFullNodesOrValidators(t, 0, func(t *testing.T, node e2e.Node) {
		if !node.EnableCompanionPruning {
			return
		}

		grpcClient, status, cleanup := getGRPCPrivilegedClientForTesting(t, node)
		defer cleanup()

		err := grpcClient.SetBlockIndexerRetainHeight(ctx, uint64(status.SyncInfo.LatestBlockHeight)-1)
		require.NoError(t, err, "Unexpected error for SetTxIndexerRetainHeight")

		height, err := grpcClient.GetBlockIndexerRetainHeight(ctx)
		require.NoError(t, err, "Unexpected error for GetTxIndexerRetainHeight")
		require.Equal(t, height, uint64(status.SyncInfo.LatestBlockHeight)-1)
	})
}

func getGRPCPrivilegedClientForTesting(t *testing.T, node e2e.Node) (privileged.Client, *coretypes.ResultStatus, func()) {
	ctx, ctxCancel := context.WithTimeout(context.Background(), time.Minute)

	grpcClient, err := node.GRPCPrivilegedClient(ctx)
	require.NoError(t, err)

	client, err := node.Client()
	require.NoError(t, err)

	status, err := client.Status(ctx)
	require.NoError(t, err)

	return grpcClient, status, func() {
		ctxCancel()
		err := grpcClient.Close()
		require.NoError(t, err)
	}
}
