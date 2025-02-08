package kcp

import (
	"io"
	"testing"
	"time"

	na "github.com/cometbft/cometbft/p2p/netaddr"
	"github.com/stretchr/testify/require"
)

func TestKCPTransportBasics(t *testing.T) {
	// Create transport with options
	opts := &Options{
		DataShards:    2,
		ParityShards:  1,
		MaxWindowSize: 32768,
		ReadTimeout:   time.Second,
		WriteTimeout:  time.Second,
	}

	transport, err := NewTransport(opts)
	require.NoError(t, err)

	// Listen on a random port
	err = transport.Listen("127.0.0.1:0")
	require.NoError(t, err)

	// Get the assigned address
	addr := transport.NetAddr()

	// Try to connect
	clientTransport, err := NewTransport(opts)
	require.NoError(t, err)

	conn, err := clientTransport.Dial(addr)
	require.NoError(t, err)

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
	require.NoError(t, clientTransport.Close())
}

func TestKCPTransportConcurrent(t *testing.T) {
	transport, err := NewTransport(nil)
	require.NoError(t, err)

	err = transport.Listen("127.0.0.1:0")
	require.NoError(t, err)

	addr := transport.NetAddr()

	// Launch multiple concurrent connections
	const numConns = 10
	done := make(chan struct{})

	for i := 0; i < numConns; i++ {
		go func() {
			clientTransport, err := NewTransport(nil)
			require.NoError(t, err)

			conn, err := clientTransport.Dial(addr)
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

		go func(conn *Conn) {
			buf := make([]byte, 1024)
			hstream := conn.HandshakeStream()
			_, err := io.ReadFull(hstream, buf[:9]) // len("test data") = 9
			require.NoError(t, err)
			require.NoError(t, conn.Close("done"))
		}(serverConn.(*Conn))
	}

	// Wait for all clients to finish
	for i := 0; i < numConns; i++ {
		<-done
	}

	require.NoError(t, transport.Close())
}

func TestKCPTransportError(t *testing.T) {
	transport, err := NewTransport(nil)
	require.NoError(t, err)

	// Try to accept before listening
	_, _, err = transport.Accept()
	require.Equal(t, ErrTransportNotListening, err)

	// Try to listen on invalid address
	err = transport.Listen("invalid-addr")
	require.Error(t, err)

	// Try to dial invalid address
	invalidAddr := na.NetAddr{}
	_, err = transport.Dial(invalidAddr)
	require.Error(t, err)

	require.NoError(t, transport.Close())
}
