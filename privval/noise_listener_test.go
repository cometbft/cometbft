package privval

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/cometbft/cometbft/crypto/ed25519"
	"github.com/cometbft/cometbft/lp2p"
	"github.com/libp2p/go-libp2p/core/peer"
	libp2pnoise "github.com/libp2p/go-libp2p/p2p/security/noise"
	"github.com/stretchr/testify/require"
)

func dialNoise(t *testing.T, addr string, clientKey ed25519.PrivKey, serverPeer peer.ID) (net.Conn, error) {
	t.Helper()
	lpk, err := lp2p.PrivateKeyFromCosmosKey(clientKey)
	require.NoError(t, err)
	tr, err := libp2pnoise.New(libp2pnoise.ID, lpk, nil)
	require.NoError(t, err)
	raw, err := net.Dial("tcp", addr)
	require.NoError(t, err)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	return tr.SecureOutbound(ctx, raw, serverPeer)
}

func TestNoiseListenerAcceptsAllowlistedPeer(t *testing.T) {
	serverKey := ed25519.GenPrivKey()
	clientKey := ed25519.GenPrivKey()
	clientPeer, err := lp2p.IDFromPrivateKey(clientKey)
	require.NoError(t, err)
	serverPeer, err := lp2p.IDFromPrivateKey(serverKey)
	require.NoError(t, err)

	tcpLn, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	nl, err := NewNoiseListener(tcpLn, serverKey, []peer.ID{clientPeer})
	require.NoError(t, err)
	defer nl.Close()

	accepted := make(chan net.Conn, 1)
	go func() {
		c, aerr := nl.Accept()
		if aerr == nil {
			accepted <- c
		} else {
			accepted <- nil
		}
	}()

	cconn, err := dialNoise(t, tcpLn.Addr().String(), clientKey, serverPeer)
	require.NoError(t, err)
	defer cconn.Close()

	select {
	case c := <-accepted:
		require.NotNil(t, c)
		c.Close()
	case <-time.After(3 * time.Second):
		t.Fatal("listener did not accept allowlisted peer")
	}
}

func TestNoiseListenerRejectsNonAllowlistedPeer(t *testing.T) {
	serverKey := ed25519.GenPrivKey()
	clientKey := ed25519.GenPrivKey()
	otherPeer, err := lp2p.IDFromPrivateKey(ed25519.GenPrivKey())
	require.NoError(t, err)
	serverPeer, err := lp2p.IDFromPrivateKey(serverKey)
	require.NoError(t, err)

	tcpLn, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	nl, err := NewNoiseListener(tcpLn, serverKey, []peer.ID{otherPeer})
	require.NoError(t, err)
	defer nl.Close()

	errCh := make(chan error, 1)
	go func() {
		_, aerr := nl.Accept()
		errCh <- aerr
	}()

	_, _ = dialNoise(t, tcpLn.Addr().String(), clientKey, serverPeer)

	select {
	case aerr := <-errCh:
		require.Error(t, aerr)
	case <-time.After(3 * time.Second):
		t.Fatal("listener did not reject non-allowlisted peer")
	}
}

func TestParseNoiseAddr(t *testing.T) {
	pid, err := lp2p.IDFromPrivateKey(ed25519.GenPrivKey())
	require.NoError(t, err)

	gotPID, hostport, err := ParseNoiseAddr("noise://" + pid.String() + "@1.2.3.4:26659")
	require.NoError(t, err)
	require.Equal(t, pid, gotPID)
	require.Equal(t, "1.2.3.4:26659", hostport)

	_, _, err = ParseNoiseAddr("noise://1.2.3.4:26659")
	require.Error(t, err)
	_, _, err = ParseNoiseAddr("tcp://1.2.3.4:26659")
	require.Error(t, err)
}
