package e2e_test

import (
	"context"
	v1 "github.com/cometbft/cometbft/proto/tendermint/services/block/v1"
	cmtversion "github.com/cometbft/cometbft/proto/tendermint/version"
	"testing"
	"time"

	e2e "github.com/cometbft/cometbft/test/e2e/pkg"
	"github.com/cometbft/cometbft/types"
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

func TestGRPC_Block(t *testing.T) {
	testNode(t, func(t *testing.T, node e2e.Node) {
		if node.Mode != e2e.ModeFull && node.Mode != e2e.ModeValidator {
			return
		}

		ctx, ctxCancel := context.WithTimeout(context.Background(), time.Minute)
		defer ctxCancel()
		client, err := node.GRPCClient(ctx)
		require.NoError(t, err)

		res, err := client.GetBlock(ctx, v1.GetBlockRequest{
			Height: 1,
		})

		require.NoError(t, err)

		header := types.Header{
			Version:            cmtversion.Consensus{},
			ChainID:            "chain-123",
			Height:             1,
			Time:               time.Time{},
			LastBlockID:        types.BlockID{},
			LastCommitHash:     nil,
			DataHash:           nil,
			ValidatorsHash:     nil,
			NextValidatorsHash: nil,
			ConsensusHash:      nil,
			AppHash:            nil,
			LastResultsHash:    nil,
			EvidenceHash:       nil,
			ProposerAddress:    nil,
		}
		block := types.Block{
			Header:     header,
			Data:       types.Data{},
			Evidence:   types.EvidenceData{},
			LastCommit: nil,
		}

		require.Equal(t, block.Height, res.Height)
		require.Equal(t, block.ChainID, res.ChainID)
	})
}
