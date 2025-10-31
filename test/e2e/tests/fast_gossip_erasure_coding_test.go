package e2e_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	e2e "github.com/cometbft/cometbft/test/e2e/pkg"
)

// TestFastGossipErasureCoding_BlockProduction tests that blocks are produced correctly
// when both fast block gossip and erasure coding are enabled
func TestFastGossipErasureCoding_BlockProduction(t *testing.T) {
	testnet := loadTestnet(t)

	// Skip if this testnet doesn't have both features enabled
	if testnet.BlockPartEncoding != "reed_solomon" || !testnet.FastBlockGossip {
		t.Skipf("Skipping: testnet does not have both erasure coding and fast block gossip enabled (encoding=%s, fast_gossip=%v)",
			testnet.BlockPartEncoding, testnet.FastBlockGossip)
	}

	testNode(t, func(t *testing.T, node e2e.Node) {
		if node.Mode == e2e.ModeSeed {
			return
		}

		client, err := node.Client()
		require.NoError(t, err)

		// Wait for the node to produce/receive multiple blocks
		require.Eventuallyf(t, func() bool {
			status, err := client.Status(ctx)
			if err != nil {
				t.Logf("error getting status: %v", err)
				return false
			}
			return status.SyncInfo.LatestBlockHeight >= node.Testnet.InitialHeight+10
		}, 30*time.Second, 500*time.Millisecond, "node %s failed to reach height %d", node.Name, node.Testnet.InitialHeight+10)

		// Verify blocks are valid
		status, err := client.Status(ctx)
		require.NoError(t, err)

		for h := node.Testnet.InitialHeight; h <= status.SyncInfo.LatestBlockHeight; h++ {
			block, err := client.Block(ctx, &h)
			require.NoError(t, err, "failed to get block at height %d", h)
			require.NotNil(t, block.Block, "block at height %d is nil", h)
			require.NoError(t, block.Block.ValidateBasic(), "block at height %d failed validation", h)
		}
	})
}

// TestFastGossipErasureCoding_Consensus verifies that consensus progresses correctly
// and all nodes agree on the same blocks
func TestFastGossipErasureCoding_Consensus(t *testing.T) {
	testnet := loadTestnet(t)

	// Skip if this testnet doesn't have both features enabled
	if testnet.BlockPartEncoding != "reed_solomon" || !testnet.FastBlockGossip {
		t.Skipf("Skipping: testnet does not have both erasure coding and fast block gossip enabled (encoding=%s, fast_gossip=%v)",
			testnet.BlockPartEncoding, testnet.FastBlockGossip)
	}

	blocks := fetchBlockChain(t)

	require.Greater(t, len(blocks), 10,
		"testnet should have produced at least 10 blocks")

	// Verify all nodes have the same blocks
	testNode(t, func(t *testing.T, node e2e.Node) {
		if node.Mode == e2e.ModeSeed {
			return
		}

		client, err := node.Client()
		require.NoError(t, err)
		status, err := client.Status(ctx)
		require.NoError(t, err)

		first := status.SyncInfo.EarliestBlockHeight
		last := status.SyncInfo.LatestBlockHeight

		for _, expectedBlock := range blocks {
			if expectedBlock.Height < first || expectedBlock.Height > last {
				continue
			}

			actualBlock, err := client.Block(ctx, &expectedBlock.Height)
			require.NoError(t, err, "node %s failed to get block at height %d",
				node.Name, expectedBlock.Height)

			require.Equal(t, expectedBlock.Hash(), actualBlock.Block.Hash(),
				"node %s has different block hash at height %d", node.Name, expectedBlock.Height)
		}
	})

	// Verify block production rate is reasonable
	if len(blocks) >= 2 {
		firstBlock := blocks[0]
		lastBlock := blocks[len(blocks)-1]
		duration := lastBlock.Time.Sub(firstBlock.Time)
		blockCount := lastBlock.Height - firstBlock.Height

		if blockCount > 0 {
			avgBlockTime := duration / time.Duration(blockCount)
			t.Logf("Average block time with fast gossip + erasure coding: %v", avgBlockTime)

			// Block time should be reasonable (not slower than 3 seconds per block)
			assert.Less(t, avgBlockTime, 3*time.Second,
				"block production is too slow with fast gossip enabled")
		}
	}

	t.Logf("Successfully verified %d blocks across %d nodes with fast gossip + erasure coding",
		len(blocks), len(testnet.Nodes))
}

// TestFastGossipOnly_BlockProduction tests that blocks are produced correctly
// with fast gossip enabled but without erasure coding
func TestFastGossipOnly_BlockProduction(t *testing.T) {
	testnet := loadTestnet(t)

	// Skip if this testnet doesn't have only fast gossip enabled
	if testnet.BlockPartEncoding != "none" && testnet.BlockPartEncoding != "" || !testnet.FastBlockGossip {
		t.Skipf("Skipping: testnet does not have only fast gossip enabled (encoding=%s, fast_gossip=%v)",
			testnet.BlockPartEncoding, testnet.FastBlockGossip)
	}

	testNode(t, func(t *testing.T, node e2e.Node) {
		if node.Mode == e2e.ModeSeed {
			return
		}

		client, err := node.Client()
		require.NoError(t, err)

		// Wait for the node to produce/receive multiple blocks
		require.Eventuallyf(t, func() bool {
			status, err := client.Status(ctx)
			if err != nil {
				return false
			}
			return status.SyncInfo.LatestBlockHeight >= node.Testnet.InitialHeight+10
		}, 30*time.Second, 500*time.Millisecond,
			"node %s failed to reach height %d", node.Name, node.Testnet.InitialHeight+10)

		// Verify blocks are valid
		status, err := client.Status(ctx)
		require.NoError(t, err)

		for h := node.Testnet.InitialHeight; h <= status.SyncInfo.LatestBlockHeight; h++ {
			block, err := client.Block(ctx, &h)
			require.NoError(t, err, "failed to get block at height %d", h)
			require.NotNil(t, block.Block)
			require.NoError(t, block.Block.ValidateBasic())
		}
	})
}

// TestFastGossipOnly_Consensus verifies consensus with only fast gossip enabled
func TestFastGossipOnly_Consensus(t *testing.T) {
	testnet := loadTestnet(t)

	// Skip if this testnet doesn't have only fast gossip enabled
	if testnet.BlockPartEncoding != "none" && testnet.BlockPartEncoding != "" || !testnet.FastBlockGossip {
		t.Skipf("Skipping: testnet does not have only fast gossip enabled (encoding=%s, fast_gossip=%v)",
			testnet.BlockPartEncoding, testnet.FastBlockGossip)
	}

	blocks := fetchBlockChain(t)

	require.Greater(t, len(blocks), 10,
		"testnet should have produced at least 10 blocks")

	// Verify all nodes agree on blocks
	testNode(t, func(t *testing.T, node e2e.Node) {
		if node.Mode == e2e.ModeSeed {
			return
		}

		client, err := node.Client()
		require.NoError(t, err)
		status, err := client.Status(ctx)
		require.NoError(t, err)

		first := status.SyncInfo.EarliestBlockHeight
		last := status.SyncInfo.LatestBlockHeight

		for _, expectedBlock := range blocks {
			if expectedBlock.Height < first || expectedBlock.Height > last {
				continue
			}

			actualBlock, err := client.Block(ctx, &expectedBlock.Height)
			require.NoError(t, err)

			require.Equal(t, expectedBlock.Hash(), actualBlock.Block.Hash(), "node %s has different block at height %d", node.Name, expectedBlock.Height)
		}
	})
}

// TestErasureCodingOnly_BlockProduction tests that blocks are produced correctly
// with erasure coding enabled but without fast gossip
func TestErasureCodingOnly_BlockProduction(t *testing.T) {
	testnet := loadTestnet(t)

	// Skip if this testnet doesn't have only erasure coding enabled
	if testnet.BlockPartEncoding != "reed_solomon" || testnet.FastBlockGossip {
		t.Skipf("Skipping: testnet does not have only erasure coding enabled (encoding=%s, fast_gossip=%v)",
			testnet.BlockPartEncoding, testnet.FastBlockGossip)
	}

	testNode(t, func(t *testing.T, node e2e.Node) {
		if node.Mode == e2e.ModeSeed {
			return
		}

		client, err := node.Client()
		require.NoError(t, err)

		// Wait for the node to produce/receive multiple blocks
		require.Eventuallyf(t, func() bool {
			status, err := client.Status(ctx)
			if err != nil {
				return false
			}
			return status.SyncInfo.LatestBlockHeight >= node.Testnet.InitialHeight+10
		}, 30*time.Second, 500*time.Millisecond,
			"node %s failed to reach height %d", node.Name, node.Testnet.InitialHeight+10)

		// Verify blocks are valid
		status, err := client.Status(ctx)
		require.NoError(t, err)

		for h := node.Testnet.InitialHeight; h <= status.SyncInfo.LatestBlockHeight; h++ {
			block, err := client.Block(ctx, &h)
			require.NoError(t, err, "failed to get block at height %d", h)
			require.NotNil(t, block.Block)
			require.NoError(t, block.Block.ValidateBasic())
		}
	})
}

// TestErasureCodingOnly_Consensus verifies consensus with only erasure coding enabled
func TestErasureCodingOnly_Consensus(t *testing.T) {
	testnet := loadTestnet(t)

	// Skip if this testnet doesn't have only erasure coding enabled
	if testnet.BlockPartEncoding != "reed_solomon" || testnet.FastBlockGossip {
		t.Skipf("Skipping: testnet does not have only erasure coding enabled (encoding=%s, fast_gossip=%v)",
			testnet.BlockPartEncoding, testnet.FastBlockGossip)
	}

	blocks := fetchBlockChain(t)

	require.Greater(t, len(blocks), 10,
		"testnet should have produced at least 10 blocks")

	// Verify all nodes agree on blocks
	testNode(t, func(t *testing.T, node e2e.Node) {
		if node.Mode == e2e.ModeSeed {
			return
		}

		client, err := node.Client()
		require.NoError(t, err)
		status, err := client.Status(ctx)
		require.NoError(t, err)

		first := status.SyncInfo.EarliestBlockHeight
		last := status.SyncInfo.LatestBlockHeight

		for _, expectedBlock := range blocks {
			if expectedBlock.Height < first || expectedBlock.Height > last {
				continue
			}

			actualBlock, err := client.Block(ctx, &expectedBlock.Height)
			require.NoError(t, err)

			require.Equal(t, expectedBlock.Hash(), actualBlock.Block.Hash(), "node %s has different block at height %d", node.Name, expectedBlock.Height)
		}
	})
}
