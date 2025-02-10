package kcp_test

import (
	"fmt"
	"io"
	"testing"

	"github.com/cometbft/cometbft/crypto/ed25519"
	"github.com/cometbft/cometbft/p2p/internal/nodekey"
	na "github.com/cometbft/cometbft/p2p/netaddr"
	"github.com/cometbft/cometbft/p2p/transport/kcp"
	"github.com/stretchr/testify/require"
)

func testAddr(t *testing.T) *na.NetAddr {
	// Create a random node ID for testing
	privKey := ed25519.GenPrivKey()
	nodeID := nodekey.PubKeyToID(privKey.PubKey())

	// Create address with ID
	addr, err := na.NewFromString(fmt.Sprintf("%s@127.0.0.1:0", nodeID))
	require.NoError(t, err)
	return addr
}

func TestKCPTransportBasics(t *testing.T) {
	privKey := ed25519.GenPrivKey()
	nodeID := nodekey.PubKeyToID(privKey.PubKey())

	opts := &kcp.Options{
		DataShards:    2,
		ParityShards:  1,
		MaxWindowSize: 32768,
	}

	transport, err := kcp.NewTransport(opts)
	require.NoError(t, err)

	// Create address without DNS lookup
	addr, err := na.NewFromString(fmt.Sprintf("%s@127.0.0.1:0", nodeID))
	require.NoError(t, err)

	err = transport.Listen(*addr)
	require.NoError(t, err)

	// Use the actual listening address
	netAddr := transport.NetAddr()
	addr = &netAddr

	conn, err := transport.Dial(*addr)
	require.NoError(t, err)
	require.NotNil(t, conn)

	// Write some data using the handshake stream
	testData := []byte("hello world")
	hstream := conn.HandshakeStream()
	n, err := hstream.Write(testData)
	require.NoError(t, err)
	require.Equal(t, len(testData), n)

	// Accept the connection on the server side
	serverConn, _, err := transport.Accept()
	require.NoError(t, err)

	// Read the data from the handshake stream
	buf := make([]byte, len(testData))
	hstream = serverConn.HandshakeStream()
	n, err = io.ReadFull(hstream, buf)
	require.NoError(t, err)
	require.Equal(t, len(testData), n)
	require.Equal(t, testData, buf)

	// Test opening a stream
	clientStream, err := conn.OpenStream(1, nil)
	require.NoError(t, err)

	streamData := []byte("stream test")
	n, err = clientStream.Write(streamData)
	require.NoError(t, err)
	require.Equal(t, len(streamData), n)

	// Close connections
	require.NoError(t, conn.Close("test done"))
	require.NoError(t, serverConn.Close("test done"))
	require.NoError(t, transport.Close())
}

func TestKCPTransportConcurrent(t *testing.T) {
	transport, err := kcp.NewTransport(nil)
	require.NoError(t, err)

	addr := testAddr(t)
	err = transport.Listen(*addr)
	require.NoError(t, err)

	// Launch multiple concurrent connections
	const numConns = 10
	done := make(chan struct{})

	for i := 0; i < numConns; i++ {
		go func() {
			clientTransport, err := kcp.NewTransport(nil)
			require.NoError(t, err)

			conn, err := clientTransport.Dial(*addr)
			require.NoError(t, err)

			data := []byte("test data")
			hstream := conn.HandshakeStream()
			_, err = hstream.Write(data)
			require.NoError(t, err)

			require.NoError(t, conn.Close("done"))
			require.NoError(t, clientTransport.Close())
			done <- struct{}{}
		}()
	}

	// Accept and handle all connections
	for i := 0; i < numConns; i++ {
		serverConn, _, err := transport.Accept()
		require.NoError(t, err)

		go func(conn *kcp.Conn) {
			buf := make([]byte, 1024)
			hstream := conn.HandshakeStream()
			_, err := io.ReadFull(hstream, buf[:9]) // len("test data") = 9
			require.NoError(t, err)
			require.NoError(t, conn.Close("done"))
		}(serverConn.(*kcp.Conn))
	}

	// Wait for all clients to finish
	for i := 0; i < numConns; i++ {
		<-done
	}

	require.NoError(t, transport.Close())
}

func TestKCPTransportError(t *testing.T) {
	transport, err := kcp.NewTransport(nil)
	require.NoError(t, err)

	// Try to accept before listening
	_, _, err = transport.Accept()
	require.Equal(t, kcp.ErrTransportNotListening, err)

	// Try to listen on invalid address
	invalidAddr, err := na.NewFromString("deadbeef@invalid-addr")
	require.NoError(t, err)
	err = transport.Listen(*invalidAddr)
	require.Error(t, err)

	// Try to dial invalid address
	_, err = transport.Dial(*invalidAddr)
	require.Error(t, err)

	require.NoError(t, transport.Close())
}
