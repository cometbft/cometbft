package transport_test

import (
	"net"
	"testing"
	"time"

	"github.com/cometbft/cometbft/crypto/ed25519"
	"github.com/cometbft/cometbft/lp2p"
	"github.com/cometbft/cometbft/privval"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/stretchr/testify/require"

	"github.com/cometbft/cometbft/kms/internal/transport"
)

func TestNoiseDialerRoundTrip(t *testing.T) {
	serverKey := ed25519.GenPrivKey()
	clientKey := ed25519.GenPrivKey()
	serverPeer, err := lp2p.IDFromPrivateKey(serverKey)
	require.NoError(t, err)
	clientPeer, err := lp2p.IDFromPrivateKey(clientKey)
	require.NoError(t, err)

	tcpLn, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	nl, err := privval.NewNoiseListener(tcpLn, serverKey, []peer.ID{clientPeer})
	require.NoError(t, err)
	defer nl.Close()

	go func() {
		c, aerr := nl.Accept()
		if aerr == nil {
			_, _ = c.Write([]byte("hi"))
			_ = c.Close()
		}
	}()

	dial, err := transport.NoiseDialer(tcpLn.Addr().String(), clientKey, serverPeer, 2*time.Second)
	require.NoError(t, err)
	conn, err := dial()
	require.NoError(t, err)
	defer conn.Close()

	buf := make([]byte, 2)
	_ = conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	_, err = conn.Read(buf)
	require.NoError(t, err)
	require.Equal(t, "hi", string(buf))
}

func TestNoiseDialerWrongPeerRejected(t *testing.T) {
	serverKey := ed25519.GenPrivKey()
	clientKey := ed25519.GenPrivKey()
	wrongPeer, err := lp2p.IDFromPrivateKey(ed25519.GenPrivKey())
	require.NoError(t, err)

	tcpLn, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	nl, err := privval.NewNoiseListener(tcpLn, serverKey, nil) // allow any
	require.NoError(t, err)
	defer nl.Close()
	go func() { _, _ = nl.Accept() }()

	dial, err := transport.NoiseDialer(tcpLn.Addr().String(), clientKey, wrongPeer, 2*time.Second)
	require.NoError(t, err)
	_, err = dial() // must fail: server's key != wrongPeer
	require.Error(t, err)
}
