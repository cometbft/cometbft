package abcicli_test

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	abcicli "github.com/cometbft/cometbft/abci/client"
	abciserver "github.com/cometbft/cometbft/abci/server"
	"github.com/cometbft/cometbft/abci/types"
	"github.com/cometbft/cometbft/libs/log"
)

// TestGRPCResponseCallbackNoDeadlock verifies that a response callback can
// call back into the client without deadlocking.
func TestGRPCResponseCallbackNoDeadlock(t *testing.T) {
	socketFile := fmt.Sprintf("/tmp/test-%08x.sock", rand.Int31n(1<<30))
	defer os.Remove(socketFile)
	socket := fmt.Sprintf("unix://%v", socketFile)

	server := abciserver.NewGRPCServer(socket, types.NewBaseApplication())
	server.SetLogger(log.TestingLogger().With("module", "abci-server"))
	require.NoError(t, server.Start())
	t.Cleanup(func() { _ = server.Stop() })

	c := abcicli.NewGRPCClient(socket, true)
	require.NoError(t, c.Start())
	t.Cleanup(func() { _ = c.Stop() })

	var once sync.Once
	done := make(chan struct{})
	c.SetResponseCallback(func(_ *types.Request, _ *types.Response) {
		_ = c.Error() // re-enters cli.mtx; deadlocks without the fix
		once.Do(func() { close(done) })
	})

	_, err := c.CheckTxAsync(context.Background(), &types.RequestCheckTx{})
	require.NoError(t, err)

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("deadlock: response callback did not complete")
	}
}

// TestGRPCResponseCallbackSeesErrorState verifies that a response callback can
// read a non-nil Error() when the client is in an error state.
func TestGRPCResponseCallbackSeesErrorState(t *testing.T) {
	socketFile := fmt.Sprintf("/tmp/test-%08x.sock", rand.Int31n(1<<30))
	defer os.Remove(socketFile)
	socket := fmt.Sprintf("unix://%v", socketFile)

	server := abciserver.NewGRPCServer(socket, types.NewBaseApplication())
	server.SetLogger(log.TestingLogger().With("module", "abci-server"))
	require.NoError(t, server.Start())
	t.Cleanup(func() { _ = server.Stop() })

	c := abcicli.NewGRPCClient(socket, true)
	require.NoError(t, c.Start())
	t.Cleanup(func() { _ = c.Stop() })

	inCallback := make(chan struct{})
	proceed := make(chan struct{})
	cbErrCh := make(chan error, 1)
	var once sync.Once

	c.SetResponseCallback(func(_ *types.Request, _ *types.Response) {
		once.Do(func() {
			close(inCallback)
			<-proceed
			cbErrCh <- c.Error()
		})
	})

	_, err := c.CheckTxAsync(context.Background(), &types.RequestCheckTx{})
	require.NoError(t, err)

	select {
	case <-inCallback:
	case <-time.After(5 * time.Second):
		t.Fatal("callback did not start")
	}

	type errorSetter interface{ StopForError(error) }
	injected := errors.New("injected test error")
	c.(errorSetter).StopForError(injected)

	close(proceed)

	select {
	case cbErr := <-cbErrCh:
		require.ErrorIs(t, cbErr, injected)
	case <-time.After(5 * time.Second):
		t.Fatal("callback did not complete")
	}
}

func TestGRPC(t *testing.T) {
	app := types.NewBaseApplication()
	numCheckTxs := 2000
	socketFile := fmt.Sprintf("/tmp/test-%08x.sock", rand.Int31n(1<<30))
	defer os.Remove(socketFile)
	socket := fmt.Sprintf("unix://%v", socketFile)

	// Start the listener
	server := abciserver.NewGRPCServer(socket, app)
	server.SetLogger(log.TestingLogger().With("module", "abci-server"))
	err := server.Start()
	require.NoError(t, err)

	t.Cleanup(func() {
		if err := server.Stop(); err != nil {
			t.Error(err)
		}
	})

	// Connect to the socket
	conn, err := grpc.NewClient(socket, grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err)

	t.Cleanup(func() {
		if err := conn.Close(); err != nil {
			t.Error(err)
		}
	})

	client := types.NewABCIClient(conn)

	// Write requests
	for counter := 0; counter < numCheckTxs; counter++ {
		// Send request
		response, err := client.CheckTx(context.Background(), &types.RequestCheckTx{Tx: []byte("test")})
		require.NoError(t, err)
		if response.Code != 0 {
			t.Error("CheckTx failed with ret_code", response.Code)
		}
		t.Log("response", counter)
	}
}
