package abcicli_test

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	abcicli "github.com/cometbft/cometbft/abci/client"
	"github.com/cometbft/cometbft/abci/server"
	"github.com/cometbft/cometbft/abci/types"
	cmtrand "github.com/cometbft/cometbft/libs/rand"
	"github.com/cometbft/cometbft/libs/service"
)

// TestCalls tests basic synchronous ABCI calls through the socket client.
// It verifies that the client can successfully communicate with the server
// and receive responses for simple operations like Echo.
func TestCalls(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	// Use a basic ABCI application for simple testing
	app := types.BaseApplication{}

	_, c := setupClientServer(t, app)

	// Test Echo call in a goroutine to verify async behavior
	resp := make(chan error, 1)
	go func() {
		res, err := c.Echo(ctx, "hello")
		require.NoError(t, err)
		require.NotNil(t, res)
		// Check for any client errors after the call
		resp <- c.Error()
	}()

	select {
	case <-time.After(time.Second):
		require.Fail(t, "No response arrived")
	case err, ok := <-resp:
		require.True(t, ok, "Must not close channel")
		assert.NoError(t, err, "This should return success")
	}
}

// TestHangingAsyncCalls tests the behavior of async calls when the server
// is terminated while a request is in progress. This verifies proper error
// handling and connection cleanup.
func TestHangingAsyncCalls(t *testing.T) {
	// Use a slow application that takes time to respond
	app := slowApp{}

	s, c := setupClientServer(t, app)

	resp := make(chan error, 1)
	go func() {
		// Start an async CheckTx call that will take time to complete
		reqres, err := c.CheckTxAsync(context.Background(), &types.RequestCheckTx{})
		require.NoError(t, err)
		// Wait for the request to be sent over the socket, but before
		// the slow server responds (server takes 1 second to respond)
		time.Sleep(50 * time.Millisecond)
		// Terminate the server to break the connection while request is pending
		err = s.Stop()
		require.NoError(t, err)

		// Wait for the CheckTx response (should fail due to connection loss)
		reqres.Wait()
		resp <- c.Error()
	}()

	select {
	case <-time.After(time.Second):
		require.Fail(t, "No response arrived")
	case err, ok := <-resp:
		require.True(t, ok, "Must not close channel")
		assert.Error(t, err, "We should get EOF error")
	}
}

// TestBulk tests the socket client with a large number of transactions
// in a single FinalizeBlock call. This verifies performance and memory
// handling with bulk operations.
func TestBulk(t *testing.T) {
	const numTxs = 700000
	// Use a Unix domain socket instead of TCP port for better performance
	socketFile := fmt.Sprintf("test-%08x.sock", rand.Int31n(1<<30))
	defer os.Remove(socketFile)
	socket := fmt.Sprintf("unix://%v", socketFile)
	// Use a base application that can handle bulk transactions
	app := types.NewBaseApplication()
	// Start the ABCI server on the Unix domain socket
	server := server.NewSocketServer(socket, app)
	t.Cleanup(func() {
		if err := server.Stop(); err != nil {
			t.Log(err)
		}
	})
	err := server.Start()
	require.NoError(t, err)

	// Create and connect the socket client (mustConnect=false for this test)
	client := abcicli.NewSocketClient(socket, false)

	t.Cleanup(func() {
		if err := client.Stop(); err != nil {
			t.Log(err)
		}
	})

	err = client.Start()
	require.NoError(t, err)

	// Construct a FinalizeBlock request with many transactions
	rfb := &types.RequestFinalizeBlock{Txs: make([][]byte, numTxs)}
	for counter := 0; counter < numTxs; counter++ {
		rfb.Txs[counter] = []byte("test")
	}
	// Send the bulk request and verify all transactions are processed
	res, err := client.FinalizeBlock(context.Background(), rfb)
	require.NoError(t, err)
	require.Equal(t, numTxs, len(res.TxResults), "Number of txs doesn't match")
	// Verify all transactions were accepted (code 0 means success)
	for _, tx := range res.TxResults {
		require.Equal(t, uint32(0), tx.Code, "Tx failed")
	}

	// Send a final flush to ensure all data is transmitted
	err = client.Flush(context.Background())
	require.NoError(t, err)
}

// setupClientServer creates a socket server and client for testing.
// It returns both the server service and the connected client.
// The server uses a random port to avoid conflicts with other tests.
func setupClientServer(t *testing.T, app types.Application) (
	service.Service, abcicli.Client,
) {
	t.Helper()

	// Use a random port between 20k and 30k to avoid conflicts
	port := 20000 + cmtrand.Int32()%10000
	addr := fmt.Sprintf("localhost:%d", port)

	// Start the ABCI server
	s := server.NewSocketServer(addr, app)
	err := s.Start()
	require.NoError(t, err)

	t.Cleanup(func() {
		if err := s.Stop(); err != nil {
			t.Log(err)
		}
	})

	// Create and start the client (mustConnect=true for reliable testing)
	c := abcicli.NewSocketClient(addr, true)
	err = c.Start()
	require.NoError(t, err)

	t.Cleanup(func() {
		if err := c.Stop(); err != nil {
			t.Log(err)
		}
	})

	return s, c
}

// slowApp is a test application that deliberately takes time to respond.
// This is used to test timeout and connection handling scenarios.
type slowApp struct {
	types.BaseApplication
}

// CheckTx implements a slow CheckTx that takes 1 second to respond.
// This allows testing of async operations and connection timeouts.
func (slowApp) CheckTx(context.Context, *types.RequestCheckTx) (*types.ResponseCheckTx, error) {
	time.Sleep(time.Second)
	return &types.ResponseCheckTx{}, nil
}

// TestCallbackInvokedWhenSetLate tests that callbacks are properly invoked
// even when set after the client has already completed the call to the app.
// This test verifies the callback mechanism works correctly for late-set callbacks.
// NOTE: This test currently relies on callbacks being allowed to be invoked
// multiple times if set multiple times.
func TestCallbackInvokedWhenSetLate(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Use a blocked application that waits for a signal before responding
	wg := &sync.WaitGroup{}
	wg.Add(1)
	app := blockedABCIApplication{
		wg: wg,
	}
	_, c := setupClientServer(t, app)
	reqRes, err := c.CheckTxAsync(ctx, &types.RequestCheckTx{})
	require.NoError(t, err)

	// Set up a callback that signals completion
	done := make(chan struct{})
	cb := func(_ *types.Response) {
		close(done)
	}
	reqRes.SetCallback(cb)
	// Unblock the application to allow it to respond
	app.wg.Done()
	// Wait for the first callback to be invoked
	<-done

	// Set a second callback and verify it gets invoked immediately
	// since the response is already available
	var called bool
	cb = func(_ *types.Response) {
		called = true
	}
	reqRes.SetCallback(cb)
	require.True(t, called)
}

// blockedABCIApplication is a test application that blocks on a WaitGroup
// before responding to CheckTx calls. This allows precise control over
// when responses are sent for testing callback timing.
type blockedABCIApplication struct {
	wg *sync.WaitGroup
	types.BaseApplication
}

// CheckTxAsync blocks until the WaitGroup is signaled, then calls the
// regular CheckTx method. This allows testing of callback timing scenarios.
func (b blockedABCIApplication) CheckTxAsync(ctx context.Context, r *types.RequestCheckTx) (*types.ResponseCheckTx, error) {
	b.wg.Wait()
	return b.CheckTx(ctx, r)
}

// TestCallbackInvokedWhenSetEarly tests that callbacks are properly invoked
// when set before the client completes the call to the app. This verifies
// the callback mechanism works correctly for early-set callbacks.
func TestCallbackInvokedWhenSetEarly(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Use a blocked application that waits for a signal before responding
	wg := &sync.WaitGroup{}
	wg.Add(1)
	app := blockedABCIApplication{
		wg: wg,
	}
	_, c := setupClientServer(t, app)
	reqRes, err := c.CheckTxAsync(ctx, &types.RequestCheckTx{})
	require.NoError(t, err)

	// Set up a callback before unblocking the application
	done := make(chan struct{})
	cb := func(_ *types.Response) {
		close(done)
	}
	reqRes.SetCallback(cb)
	// Unblock the application to allow it to respond
	app.wg.Done()

	// Check if the callback was invoked within a reasonable time
	called := func() bool {
		select {
		case <-done:
			return true
		default:
			return false
		}
	}
	// Wait up to 1 second for the callback to be invoked
	require.Eventually(t, called, time.Second, time.Millisecond*25)
}
