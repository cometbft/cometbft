package server

import (
	"context"
	"net"
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/cometbft/cometbft/abci/types"
)

func TestNewGRPCServerWithListener(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)

	srv := NewGRPCServerWithListener(ln, types.NewBaseApplication())
	require.NoError(t, srv.Start())
	t.Cleanup(func() { require.NoError(t, srv.Stop()) })

	conn, err := grpc.NewClient(ln.Addr().String(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, conn.Close()) })

	client := types.NewABCIClient(conn)
	resp, err := client.Echo(context.Background(), &types.RequestEcho{Message: "ping"})
	require.NoError(t, err)
	require.Equal(t, "ping", resp.Message)
}
