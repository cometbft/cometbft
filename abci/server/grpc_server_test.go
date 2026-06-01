package server

import (
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
	conn.Close()
}
