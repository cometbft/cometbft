package node

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/cometbft/cometbft/abci/example/kvstore"
	"github.com/cometbft/cometbft/blocksync"
	"github.com/cometbft/cometbft/config"
	"github.com/cometbft/cometbft/crypto/ed25519"
	"github.com/cometbft/cometbft/libs/log"
	"github.com/cometbft/cometbft/p2p"
	"github.com/cometbft/cometbft/privval"
	"github.com/cometbft/cometbft/proxy"
	"github.com/cometbft/cometbft/types"
)

func TestNativeBlockServingPeerCatchesUp(t *testing.T) {
	t.Parallel()
	key := ed25519.GenPrivKey()
	genesis := &types.GenesisDoc{
		GenesisTime: time.Now(), ChainID: "block-serving-test", InitialHeight: 1,
		ConsensusParams: types.DefaultConsensusParams(),
		Validators:      []types.GenesisValidator{{PubKey: key.PubKey(), Power: 10}},
	}
	server := newBlockServingNode(t, key, genesis)
	require.IsType(t, &blocksync.Reactor{}, server.Switch().Reactor("BLOCKSYNC"))
	waitForBlockServingHeight(t, server, 5)
	require.True(t, server.Switch().Reactor("BLOCKSYNC").IsRunning())
	client := newBlockServingNode(t, ed25519.GenPrivKey(), genesis)
	require.IsType(t, &blocksync.Reactor{}, client.Switch().Reactor("BLOCKSYNC"))
	require.NoError(t, client.Switch().DialPeerWithAddress(server.Switch().NetAddress()))
	waitForBlockServingHeight(t, client, 5)
	require.Equal(t, server.BlockStore().LoadBlock(5).Hash(), client.BlockStore().LoadBlock(5).Hash())
}

func newBlockServingNode(t *testing.T, key ed25519.PrivKey, genesis *types.GenesisDoc) *Node {
	t.Helper()
	cfg := config.TestConfig().SetRoot(t.TempDir())
	cfg.RPC.ListenAddress = ""
	cfg.P2P.ListenAddress = "tcp://" + testFreeAddr(t)
	cfg.P2P.PexReactor = false
	for _, dir := range []string{"config", "data"} {
		require.NoError(t, os.MkdirAll(filepath.Join(cfg.RootDir, dir), 0o700))
	}
	pv := privval.NewFilePV(key, cfg.PrivValidatorKeyFile(), cfg.PrivValidatorStateFile())
	genesisProvider := func() (*types.GenesisDoc, error) { return genesis, nil }
	n, err := NewNodeWithContext(context.Background(), cfg, pv,
		&p2p.NodeKey{PrivKey: ed25519.GenPrivKey()}, proxy.NewLocalClientCreator(kvstore.NewInMemoryApplication()),
		genesisProvider, config.DefaultDBProvider, DefaultMetricsProvider(cfg.Instrumentation), log.NewNopLogger())
	require.NoError(t, err)
	require.NoError(t, n.Start())
	t.Cleanup(func() {
		require.NoError(t, n.Stop())
		n.Wait()
		require.False(t, n.Switch().Reactor("BLOCKSYNC").IsRunning())
	})
	return n
}

func waitForBlockServingHeight(t *testing.T, n *Node, height int64) {
	t.Helper()
	require.Eventually(t, func() bool { return n.BlockStore().Height() >= height }, 15*time.Second, 10*time.Millisecond,
		"node did not reach height %d", height)
}
