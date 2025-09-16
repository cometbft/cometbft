package abcicli_test

import (
	"fmt"
	"math/rand"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/credentials/insecure"

	"google.golang.org/grpc"

	"golang.org/x/net/context"

	abciserver "github.com/cometbft/cometbft/abci/server"
	"github.com/cometbft/cometbft/abci/types"
	"github.com/cometbft/cometbft/libs/log"
)

// TestGRPC tests the gRPC ABCI client by sending multiple CheckTx requests
// to a gRPC server and verifying the responses.
func TestGRPC(t *testing.T) {
	// Create a base ABCI application for testing
	app := types.NewBaseApplication()
	// Number of CheckTx requests to send for load testing
	numCheckTxs := 2000
	// Generate a unique socket file path to avoid conflicts
	socketFile := fmt.Sprintf("/tmp/test-%08x.sock", rand.Int31n(1<<30))
	defer os.Remove(socketFile)
	// Format socket address for Unix domain socket
	socket := fmt.Sprintf("unix://%v", socketFile)

	// Start the gRPC ABCI server
	server := abciserver.NewGRPCServer(socket, app)
	server.SetLogger(log.TestingLogger().With("module", "abci-server"))
	err := server.Start()
	require.NoError(t, err)

	t.Cleanup(func() {
		if err := server.Stop(); err != nil {
			t.Error(err)
		}
	})

	// Connect to the gRPC server using insecure credentials (for testing)
	conn, err := grpc.NewClient(socket, grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err)

	t.Cleanup(func() {
		if err := conn.Close(); err != nil {
			t.Error(err)
		}
	})

	// Create ABCI client from gRPC connection
	client := types.NewABCIClient(conn)

	// Send multiple CheckTx requests to test client performance and reliability
	for counter := 0; counter < numCheckTxs; counter++ {
		// Send CheckTx request with test transaction data
		response, err := client.CheckTx(context.Background(), &types.RequestCheckTx{Tx: []byte("test")})
		require.NoError(t, err)
		// Verify that the transaction was accepted (code 0 means success)
		if response.Code != 0 {
			t.Error("CheckTx failed with ret_code", response.Code)
		}
		// Safety check to prevent infinite loops (should never trigger)
		if counter > numCheckTxs {
			t.Fatal("Too many CheckTx responses")
		}
		t.Log("response", counter)
		// This condition will never be true since counter < numCheckTxs in the loop
		// The goroutine appears to be dead code and should be removed
		if counter == numCheckTxs {
			go func() {
				time.Sleep(time.Second * 1) // Wait for a bit to allow counter overflow
			}()
		}

	}
}
