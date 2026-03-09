package abcicli_test

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	abciserver "github.com/cometbft/cometbft/abci/server"
	"github.com/cometbft/cometbft/abci/types"
	"github.com/cometbft/cometbft/libs/log"
)

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
