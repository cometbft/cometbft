package http_test

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/cometbft/cometbft/abci/example/kvstore"
	lighthttp "github.com/cometbft/cometbft/light/provider/http"
	rpcclient "github.com/cometbft/cometbft/rpc/client"
	rpchttp "github.com/cometbft/cometbft/rpc/client/http"
	rpctest "github.com/cometbft/cometbft/rpc/test"
	"github.com/cometbft/cometbft/types"
)

func TestNewProvider(t *testing.T) {
	t.Log("Starting TestNewProvider")
	c, err := lighthttp.New("chain-test", "192.168.0.1:26657")
	require.NoError(t, err)
	t.Logf("Created new light http provider: %s", c)
	require.Equal(t, "http{http://192.168.0.1:26657}", fmt.Sprintf("%s", c))

	c, err = lighthttp.New("chain-test", "http://153.200.0.1:26657")
	require.NoError(t, err)
	t.Logf("Created new light http provider: %s", c)
	require.Equal(t, "http{http://153.200.0.1:26657}", fmt.Sprintf("%s", c))

	c, err = lighthttp.New("chain-test", "153.200.0.1")
	require.NoError(t, err)
	t.Logf("Created new light http provider: %s", c)
	require.Equal(t, "http{http://153.200.0.1}", fmt.Sprintf("%s", c))
}

func TestProvider(t *testing.T) {
	t.Log("Starting TestProvider")
	for _, path := range []string{"", "/", "/v1", "/v1/"} {
		t.Logf("Testing with path: %s", path)
		app := kvstore.NewInMemoryApplication()
		app.RetainBlocks = 10
		t.Log("Starting CometBFT node with in-memory application")
		node := rpctest.StartCometBFT(app, rpctest.RecreateConfig)

		cfg := rpctest.GetConfig()
		defer func() {
			t.Log("Cleaning up: removing root directory")
			os.RemoveAll(cfg.RootDir)
		}()
		rpcAddr := cfg.RPC.ListenAddress
		t.Logf("RPC address: %s", rpcAddr)

		genDoc, err := types.GenesisDocFromFile(cfg.GenesisFile())
		require.NoError(t, err)
		chainID := genDoc.ChainID
		t.Logf("Chain ID: %s", chainID)

		c, err := rpchttp.New(rpcAddr + path)
		require.NoError(t, err)
		t.Logf("RPC HTTP client created for address: %s", rpcAddr+path)

		p := lighthttp.NewWithClient(chainID, c)
		require.NoError(t, err)
		require.NotNil(t, p)
		t.Log("Light client created with RPC HTTP client")

		t.Log("Waiting for node to produce blocks")
		err = rpcclient.WaitForHeight(c, 10, nil)
		require.NoError(t, err)

		t.Log("Fetching the highest block")
		lb, err := p.LightBlock(context.Background(), 0)
		require.NoError(t, err)
		require.NotNil(t, lb)
		t.Logf("Fetched light block at height: %d", lb.Height)
		assert.GreaterOrEqual(t, lb.Height, int64(10))

		t.Log("Validating the highest block")
		require.NoError(t, lb.ValidateBasic(chainID))

		t.Log("Testing historical queries")
		lower := lb.Height - 3
		lb, err = p.LightBlock(context.Background(), lower)
		require.NoError(t, err)
		t.Logf("Fetched historical light block at height: %d", lb.Height)
		assert.Equal(t, lower, lb.Height)

		t.Log("Testing error handling for missing heights")
		lb, err = p.LightBlock(context.Background(), lb.Height+100000)
		require.Error(t, err)
		require.Nil(t, lb)
		t.Log("Error fetching light block at future height: ErrHeightTooHigh")

		_, err = p.LightBlock(context.Background(), 1)
		require.Error(t, err)
		require.Nil(t, lb)
		t.Log("Error fetching light block at pruned height: ErrLightBlockNotFound")

		t.Log("Stopping the full node to test error handling for no response")
		rpctest.StopCometBFT(node)
		time.Sleep(10 * time.Second)
		lb, err = p.LightBlock(context.Background(), lower+2)
		require.Error(t, err)
		require.Nil(t, lb)
		t.Log("Expected error on fetching light block after node stop: connection refused")
	}
}
