package p2p

import (
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/cometbft/cometbft/crypto/ed25519"
	"github.com/cometbft/cometbft/p2p/nodekey"
)

func TestHandshake(t *testing.T) {
	c1, c2 := net.Pipe()

	go func() {
		var (
			pk       = ed25519.GenPrivKey()
			nodeInfo = testNodeInfo(nodekey.PubKeyToID(pk.PubKey()), "c2")
		)
		_, err := handshakeOverStream(nodeInfo, c1, 20*time.Millisecond)
		if err != nil {
			panic("handshake failed: " + err.Error())
		}
	}()

	var (
		pk       = ed25519.GenPrivKey()
		nodeInfo = testNodeInfo(nodekey.PubKeyToID(pk.PubKey()), "c1")
	)

	_, err := handshakeOverStream(nodeInfo, c2, 20*time.Millisecond)
	require.NoError(t, err)
}

func TestHandshake_InvalidNodeInfo(t *testing.T) {
	c1, c2 := net.Pipe()

	go func() {
		var (
			pk       = ed25519.GenPrivKey()
			nodeInfo = testNodeInfo(nodekey.PubKeyToID(pk.PubKey()), "c2")
		)

		// modify nodeInfo to be invalid
		nodeInfo.Other.TxIndex = "invalid"

		_, err := handshakeOverStream(nodeInfo, c1, 20*time.Millisecond)
		if err != nil {
			panic("handshake failed: " + err.Error())
		}
	}()

	var (
		pk       = ed25519.GenPrivKey()
		nodeInfo = testNodeInfo(nodekey.PubKeyToID(pk.PubKey()), "c1")
	)

	_, err := handshakeOverStream(nodeInfo, c2, 20*time.Millisecond)
	require.Error(t, err)

	if e, ok := err.(ErrRejected); ok {
		if !e.IsNodeInfoInvalid() {
			t.Errorf("expected NodeInfo to be invalid, got %v", err)
		}
	} else {
		t.Errorf("expected ErrRejected, got %v", err)
	}
}

func TestTransportMultiplexRejectSelf(t *testing.T) {
	c1, c2 := net.Pipe()

	var (
		pk1       = ed25519.GenPrivKey()
		nodeInfo1 = testNodeInfo(nodekey.PubKeyToID(pk1.PubKey()), "c1")
	)

	go func() {
		nodeInfo2 := testNodeInfo(nodeInfo1.ID(), "c2")

		_, err := handshakeOverStream(nodeInfo2, c1, 20*time.Millisecond)
		if err == nil {
			panic("expected handshake to fail")
		}
	}()

	_, err := handshakeOverStream(nodeInfo1, c2, 20*time.Millisecond)
	require.Error(t, err)

	if err, ok := err.(ErrRejected); ok {
		if !err.IsSelf() {
			t.Errorf("expected to reject self, got: %v", err)
		}
	} else {
		t.Errorf("expected ErrRejected, got %v", nil)
	}
}

func TestHandshake_Incompatible(t *testing.T) {
	c1, c2 := net.Pipe()

	go func() {
		var (
			pk       = ed25519.GenPrivKey()
			nodeInfo = testNodeInfo(nodekey.PubKeyToID(pk.PubKey()), "c2")
		)

		// modify nodeInfo to be incompatible
		nodeInfo.Network = "other"

		_, err := handshakeOverStream(nodeInfo, c1, 20*time.Millisecond)
		if err == nil {
			panic("expected handshake to fail")
		}
	}()

	var (
		pk       = ed25519.GenPrivKey()
		nodeInfo = testNodeInfo(nodekey.PubKeyToID(pk.PubKey()), "c1")
	)

	_, err := handshakeOverStream(nodeInfo, c2, 20*time.Millisecond)
	require.Error(t, err)

	if e, ok := err.(ErrRejected); ok {
		if !e.IsIncompatible() {
			t.Errorf("expected to reject incompatible, got %v", e)
		}
	} else {
		t.Errorf("expected ErrRejected, got %v", err)
	}
}
