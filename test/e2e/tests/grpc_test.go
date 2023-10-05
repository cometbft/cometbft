package e2e_test

import (
	"context"
	"testing"
	"time"

	e2e "github.com/cometbft/cometbft/test/e2e/pkg"
	"github.com/cometbft/cometbft/version"
	"github.com/stretchr/testify/require"

	legacy_grpc "github.com/cometbft/cometbft/rpc/grpc"
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

func TestLegacyGRPC_Ping(t *testing.T) {
	testNode(t, func(t *testing.T, node e2e.Node) {
		if node.Mode != e2e.ModeFull && node.Mode != e2e.ModeValidator {
			return
		}

		client, err := node.GRPCLegacyClient()
		require.NoError(t, err)

		_, err = client.Ping(context.Background(), &legacy_grpc.RequestPing{})
		require.NoError(t, err)
	})
}
