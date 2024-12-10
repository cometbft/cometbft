package e2e_test

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	e2e "github.com/cometbft/cometbft/test/e2e/pkg"
)

// Tests that block headers are identical across nodes where present.
func TestBlock_Header(t *testing.T) {
	t.Helper()
	blocks := fetchBlockChain(t)
	testNode(t, func(t *testing.T, node e2e.Node) {
		t.Helper()
		if node.Mode == e2e.ModeSeed || node.EnableCompanionPruning {
			return
		}

		client, err := node.Client()
		require.NoError(t, err)
		status, err := client.Status(ctx)
		require.NoError(t, err)

		first := status.SyncInfo.EarliestBlockHeight
		last := status.SyncInfo.LatestBlockHeight
		if node.RetainBlocks > 0 {
			// This was done in case pruning is activated.
			// As it happens in the background this lowers the chances
			// that the block at height=first will be pruned by the time we test
			// this. If this test starts to fail often, it is worth revisiting this logic.
			// To reproduce this failure locally, it is advised to set the storage.pruning.interval
			// to 1s instead of 10s.
			first += int64(node.RetainBlocks) // avoid race conditions with block pruning
		}

		for _, block := range blocks {
			if block.Header.Height < first {
				continue
			}
			if block.Header.Height > last {
				break
			}
			resp, err := client.Block(ctx, &block.Header.Height)
			require.NoError(t, err)

			require.Equal(t, block, resp.Block,
				"block mismatch for height %d", block.Header.Height)

			require.NoError(t, resp.Block.ValidateBasic(),
				"block at height %d is invalid", block.Header.Height)
		}
	})
}

// Tests that the node configured to prune are actually pruning.
func TestBlock_Pruning(t *testing.T) {
	t.Helper()
	testNode(t, func(t *testing.T, node e2e.Node) {
		t.Helper()
		// We do not run this test on stateless nodes or nodes with data
		// companion-related pruning enabled or nodes not pruning.
		if node.Stateless() || node.EnableCompanionPruning || node.RetainBlocks == 0 {
			return
		}
		client, err := node.Client()
		require.NoError(t, err)
		status, err := client.Status(ctx)
		require.NoError(t, err)
		first0 := status.SyncInfo.EarliestBlockHeight

		require.Eventually(t, func() bool {
			status, err := client.Status(ctx)
			require.NoError(t, err)
			first := status.SyncInfo.EarliestBlockHeight
			last := status.SyncInfo.LatestBlockHeight
			pruning := first > first0
			pruningEnough := last-first+1 < int64(node.RetainBlocks)+10 // 10 represents some leeway
			return pruning && pruningEnough
		}, 1*time.Minute, 3*time.Second, "node %v is not pruning correctly", node.Name)
	})
}

// Tests that the node contains the expected block range.
func TestBlock_Range(t *testing.T) {
	t.Helper()
	testNode(t, func(t *testing.T, node e2e.Node) {
		t.Helper()

		client, err := node.Client()
		require.NoError(t, err)
		status, err := client.Status(ctx)
		require.NoError(t, err)

		first := status.SyncInfo.EarliestBlockHeight
		last := status.SyncInfo.LatestBlockHeight

		switch {
		case node.StateSync:
			assert.Greater(t, first, node.Testnet.InitialHeight,
				"state synced nodes should not contain network's initial height")

		case node.RetainBlocks > 0 && int64(node.RetainBlocks) < (last-node.Testnet.InitialHeight+1):
			// we test nodes are actually pruning in `TestBlock_Pruning`
			assert.GreaterOrEqual(t, uint64(last-first+1), node.RetainBlocks, "node pruned more blocks than it should")

		default:
			assert.Equal(t, node.Testnet.InitialHeight, first,
				"node's first block should be network's initial height")
		}

		for h := first; h <= last; h++ {
			if h < first {
				continue
			}
			resp, err := client.Block(ctx, &(h))
			if node.RetainBlocks > 0 && !node.EnableCompanionPruning &&
				(err != nil && strings.Contains(err.Error(), " is not available, lowest height is ") ||
					resp.Block == nil) {
				// If node is pruning and doesn't return a valid block
				// compare wanted block to blockstore's base, and update `first`.
				status, err := client.Status(ctx)
				require.NoError(t, err)
				first = status.SyncInfo.EarliestBlockHeight
				if h < first {
					continue
				}
			}
			require.NoError(t, err)
			require.NotNil(t, resp)
			require.NotNil(t, resp.Block)
			assert.Equal(t, h, resp.Block.Height)
		}

		for h := node.Testnet.InitialHeight; h < first; h++ {
			_, err := client.Block(ctx, &(h))
			require.Error(t, err)
		}
	})
}

// Tests that time is monotonically increasing,
// and that blocks produced according to BFT Time follow MedianTime calculation.
func TestBlock_Time(t *testing.T) {
	t.Helper()
	blocks := fetchBlockChain(t)
	testnet := loadTestnet(t)

	// blocks are 1-indexed, i.e., blocks[0] = height 1, blocks[1] = height 2, etc.
	lastBlock := blocks[0]
	valSchedule := newValidatorSchedule(t, &testnet)
	for _, block := range blocks[1:] {
		require.Less(t, lastBlock.Time, block.Time)
		lastBlock = block

		if testnet.PbtsEnableHeight == 0 || block.Height < testnet.PbtsEnableHeight {
			expTime := block.LastCommit.MedianTime(valSchedule.Set)

			require.Equal(
				t,
				expTime,
				block.Time,
				"height=%d, valSet=%s\n%s",
				block.Height,
				valSchedule.Set,
				block.LastCommit.StringIndented("  "),
			)
		}

		valSchedule.IncreaseHeight(t, 1)
	}
}
