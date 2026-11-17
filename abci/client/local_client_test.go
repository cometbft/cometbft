package abcicli_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	abcicli "github.com/cometbft/cometbft/abci/client"
	"github.com/cometbft/cometbft/abci/types"
)

// TestLocalClientResponseCallbackNoDeadlock checks that a response callback may
// call back into the local client (like it can with the socket and gRPC
// clients) without deadlocking on the application mutex.
func TestLocalClientResponseCallbackNoDeadlock(t *testing.T) {
	c := abcicli.NewLocalClient(nil, types.NewBaseApplication())
	require.NoError(t, c.Start())
	t.Cleanup(func() { _ = c.Stop() })

	done := make(chan struct{})
	c.SetResponseCallback(func(_ *types.Request, _ *types.Response) {
		_, _ = c.Query(context.Background(), &types.RequestQuery{}) // takes the app mutex
		close(done)
	})

	go func() {
		_, err := c.CheckTxAsync(context.Background(), &types.RequestCheckTx{})
		require.NoError(t, err)
	}()

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("deadlock: response callback did not complete")
	}
}

// TestLocalClientCheckTxAsyncWithoutCallback checks that CheckTxAsync does not
// panic when no response callback is set.
func TestLocalClientCheckTxAsyncWithoutCallback(t *testing.T) {
	c := abcicli.NewLocalClient(nil, types.NewBaseApplication())
	require.NoError(t, c.Start())
	t.Cleanup(func() { _ = c.Stop() })

	rr, err := c.CheckTxAsync(context.Background(), &types.RequestCheckTx{})
	require.NoError(t, err)
	require.NotNil(t, rr.Response)
}
