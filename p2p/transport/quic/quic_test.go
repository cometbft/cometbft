package quic

import (
	"crypto/tls"
	"io"
	"testing"
	"time"

	na "github.com/cometbft/cometbft/p2p/netaddr"
	"github.com/stretchr/testify/require"
)

func generateTestTLSConfig() *tls.Config {
	return &tls.Config{
		InsecureSkipVerify: true,
		NextProtos:         []string{"quic-test"},
	}
}

func TestQUICTransportBasics(t *testing.T) {
	tlsConfig := generateTestTLSConfig()

	// Create transport with options
	opts := &Options{
		TLSConfig:          tlsConfig,
		MaxIncomingStreams: 10,
		KeepAlivePeriod:    time.Second,
		IdleTimeout:        time.Minute,
	}

	transport, err := NewTransport(opts)
	require.NoError(t, err)

	// Listen on a random port
	addr, err := na.NewFromString("127.0.0.1:0")
	require.NoError(t, err)
	err = transport.Listen(*addr)

	// Get the assigned address
	netAddr := transport.NetAddr()
	addr = &netAddr // Convert NetAddr to *NetAddr

	// Try to connect
	clientTransport, err := NewTransport(opts)
	require.NoError(t, err)

	conn, err := clientTransport.Dial(*addr)
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

	// Close connections
	require.NoError(t, conn.Close("test done"))
	require.NoError(t, serverConn.Close("test done"))
	require.NoError(t, transport.Close())
	require.NoError(t, clientTransport.Close())
}

func TestQUICTransportError(t *testing.T) {
	tlsConfig := generateTestTLSConfig()

	transport, err := NewTransport(&Options{
		TLSConfig: tlsConfig,
	})
	require.NoError(t, err)

	// Try to accept before listening
	_, _, err = transport.Accept()
	require.Equal(t, ErrTransportNotListening, err)

	// Try to listen on invalid address
	invalidAddr, err := na.NewFromString("invalid-addr")
	require.NoError(t, err)
	err = transport.Listen(*invalidAddr)

	// Try to dial invalid address
	_, err = transport.Dial(*invalidAddr)
	require.Error(t, err)

	require.NoError(t, transport.Close())
}
