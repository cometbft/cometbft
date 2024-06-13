package e2e_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	coretypes "github.com/cometbft/cometbft/rpc/core/types"
	"github.com/cometbft/cometbft/rpc/grpc/client/privileged"
	e2e "github.com/cometbft/cometbft/test/e2e/pkg"
	"github.com/cometbft/cometbft/version"
)

func TestGRPC_Version(t *testing.T) {
	t.Helper()
	testFullNodesOrValidators(t, 0, func(t *testing.T, node e2e.Node) {
		t.Helper()
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
		t.Helper()

		ctx, ctxCancel := context.WithTimeout(context.Background(), time.Minute)
		defer ctxCancel()
		gRPCClient, err := node.GRPCClient(ctx)
		require.NoError(t, err)
		defer gRPCClient.Close()

		latestHeightCh, err := gRPCClient.GetLatestHeight(ctx)
		require.NoError(t, err)

		latest, ok := <-latestHeightCh
		require.Equal(t, ok, true)
		require.Greater(t, latest.Height, int64(0))

		// Get last block and fetch it using the gRPC endpoint
		lastBlock, err := gRPCClient.GetBlockByHeight(ctx, latest.Height)

		// Last block tests
		require.NoError(t, err)
		require.NotNil(t, lastBlock.BlockID)
		require.Equal(t, lastBlock.Block.Height, latest.Height)
		require.NotNil(t, lastBlock.Block.LastCommit)
	})
}

func TestGRPC_Block_GetLatestHeight(t *testing.T) {
	t.Helper()
	testFullNodesOrValidators(t, 0, func(t *testing.T, node e2e.Node) {
		t.Helper()
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
	t.Helper()
	testFullNodesOrValidators(t, 0, func(t *testing.T, node e2e.Node) {
		t.Helper()

		ctx, ctxCancel := context.WithTimeout(context.Background(), time.Minute)
		defer ctxCancel()
		gRPCClient, err := node.GRPCClient(ctx)
		require.NoError(t, err)
		defer gRPCClient.Close()

		latestHeightCh, err := gRPCClient.GetLatestHeight(ctx)
		require.NoError(t, err)

		latest, ok := <-latestHeightCh
		require.Equal(t, ok, true)
		require.Greater(t, latest.Height, int64(0))

		// Get last block and fetch it using the gRPC endpoint
		lastBlockResults, err := gRPCClient.GetBlockResults(ctx, latest.Height)

		// Last block results tests
		require.NoError(t, err)
		require.Equal(t, lastBlockResults.Height, latest.Height)
		require.NotNil(t, lastBlockResults.AppHash)
	})
}

func TestGRPC_BlockRetainHeight(t *testing.T) {
	t.Helper()
	testFullNodesOrValidators(t, 0, func(t *testing.T, node e2e.Node) {
		t.Helper()
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
	t.Helper()
	testFullNodesOrValidators(t, 0, func(t *testing.T, node e2e.Node) {
		t.Helper()
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
		t.Helper()
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
	t.Helper()
	testFullNodesOrValidators(t, 0, func(t *testing.T, node e2e.Node) {
		t.Helper()
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
	t.Helper()
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
