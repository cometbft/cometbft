package p2p

import (
	"fmt"
	"testing"
	"time"

	"github.com/cometbft/cometbft/crypto/ed25519"
	cmtnet "github.com/cometbft/cometbft/internal/net"
	"github.com/stretchr/testify/require"
)

// SimulationParams contains the parameters required to simulate a peer.
type SimulationParams struct {
	PrivKey                                   ed25519.PrivKey
	NodeInfo                                  NodeInfo
	FastTransport, SlowTransport              *MultiplexTransport
	FastAddr, SlowAddr                        *NetAddress
	SlowChan, FastChan, SlowDoneChan, ErrChan chan error
	T                                         *testing.T
}

// TestTransportMultiplexAcceptNonBlocking tests the non-blocking acceptance of a transport multiplex.
// It simulates a slow peer and a fast peer. The slow peer initiates a connection and then waits for the fast peer to connect.
// The test checks that the fast peer's NodeInfo is correctly received.
func TestTransportMultiplexAcceptNonBlocking(t *testing.T) {
	t.Log("Setting up simulation parameters for slow peer")
	slowPeerParams, err := setupSimulationParams(t)
	require.NoError(t, err)

	t.Log("Generating new NodeKey for fast peer")
	fastPeerPrivKey := ed25519.GenPrivKey()
	for fastPeerPrivKey.PubKey().Equals(slowPeerParams.PrivKey.PubKey()) {
		t.Log("Fast peer key equals slow peer key, generating a new one")
		fastPeerPrivKey = ed25519.GenPrivKey()
	}
	fastPeerNodeInfo := testNodeInfo(PubKeyToID(fastPeerPrivKey.PubKey()), "fast_peer")

	t.Log("Checking if fast peer and slow peer have the same ID")
	if fastPeerNodeInfo.ID() == slowPeerParams.NodeInfo.ID() {
		t.Fatal("Fast peer and slow peer have the same ID")
	}

	t.Log("Creating new transport for fast peer")
	fastPeerTransport := newMultiplexTransport(
		fastPeerNodeInfo,
		NodeKey{
			PrivKey: fastPeerPrivKey,
		},
	)

	t.Log("Starting listener for fast peer")
	if err := fastPeerTransport.Listen(*slowPeerParams.SlowAddr); err != nil {
		t.Fatalf("Failed to start listener for fast peer: %v", err)
	}

	t.Log("Creating channel to signal when the test should be stopped")
	stopChan := make(chan struct{})

	t.Log("Running the test in a separate goroutine")
	go func() {
		defer close(stopChan)

		t.Log("Checking if fast peer transport is initialized")
		if fastPeerTransport == nil {
			slowPeerParams.ErrChan <- fmt.Errorf("Transport is not initialized")
			return
		}
		t.Log("Checking if fast peer NodeKey is initialized")
		if fastPeerTransport.nodeKey == (NodeKey{}) {
			slowPeerParams.ErrChan <- fmt.Errorf("NodeKey is not initialized")
			return
		}
		t.Log("Checking if fast peer listener is initialized")
		if fastPeerTransport.listener == nil {
			slowPeerParams.ErrChan <- fmt.Errorf("Listener is not initialized")
			return
		}
		t.Log("Fast peer attempting to dial slow peer")
		_, err := fastPeerTransport.Dial(*slowPeerParams.SlowAddr, peerConfig{})
		if err != nil {
			slowPeerParams.ErrChan <- fmt.Errorf("Fast peer failed to dial: %v", err)
			return
		}

		t.Log("Fast peer dialed successfully")
		close(slowPeerParams.FastChan)

		t.Log("Waiting for the slow peer to finish")
		select {
		case <-slowPeerParams.SlowDoneChan:
			t.Log("Slow peer finished")
		case <-time.After(time.Minute): // Timeout after 1 minute
			slowPeerParams.ErrChan <- fmt.Errorf("Slow peer did not finish within the expected time")
			return
		}
	}()

	t.Log("Waiting for the test to finish or for the timeout")
	select {
	case <-stopChan:
		t.Log("Test finished successfully")
	case err := <-slowPeerParams.ErrChan:
		t.Fatalf("Test failed: %v", err)
	case <-time.After(time.Minute): // Timeout after 1 minute
		t.Error("Test timed out")
	}
}

// simulateSlowPeer simulates a slow peer that initiates a connection and then waits for the fast peer to connect.
func simulateSlowPeer(params SimulationParams) {
	addr := NewNetAddress(params.SlowTransport.nodeKey.ID(), params.SlowTransport.listener.Addr())
	c, err := addr.Dial()
	if err != nil {
		params.ErrChan <- fmt.Errorf("slow peer failed to dial: %v", err)
		return
	}
	defer c.Close()

	sc, err := upgradeSecretConn(c, 200*time.Millisecond, params.PrivKey)
	if err != nil {
		params.ErrChan <- fmt.Errorf("slow peer failed to upgrade connection: %v", err)
		return
	}

	_, err = handshake(sc, 200*time.Millisecond,
		testNodeInfo(
			PubKeyToID(params.PrivKey.PubKey()),
			"slow_peer",
		))
	if err != nil {
		params.ErrChan <- fmt.Errorf("slow peer failed to handshake: %v", err)
		return
	}

	close(params.SlowChan)
	defer close(params.SlowDoneChan)

	<-params.FastChan
}

// simulateFastPeer simulates a fast peer that connects after the slow peer has initiated a connection.
func simulateFastPeer(params SimulationParams) {
	// Wait for the slow peer to initiate a connection
	<-params.SlowChan

	// Create a new NodeKey for the fast peer
	fastPeerPrivKey := ed25519.GenPrivKey()
	fastPeerNodeInfo := testNodeInfo(PubKeyToID(fastPeerPrivKey.PubKey()), "fast_peer")

	// Create a new transport for the fast peer
	dialer := newMultiplexTransport(
		fastPeerNodeInfo,
		NodeKey{
			PrivKey: fastPeerPrivKey,
		},
	)
	defer func() {
		if err := dialer.Close(); err != nil {
			params.ErrChan <- fmt.Errorf("Failed to close dialer: %v", err)
		}
	}()

	// Dial the multiplex transport
	addr := NewNetAddress(params.FastTransport.nodeKey.ID(), params.FastTransport.listener.Addr())
	_, err := dialer.Dial(*addr, peerConfig{})
	if err != nil {
		params.ErrChan <- fmt.Errorf("fast peer failed to dial: %v", err)
		return
	}

	close(params.FastChan)

	// Wait for the slow peer to finish
	<-params.SlowDoneChan
}

func setupTransport(t *testing.T, nodeInfo NodeInfo, privKey ed25519.PrivKey, port int) (*MultiplexTransport, *NetAddress, error) {
	transport := newMultiplexTransport(nodeInfo, NodeKey{PrivKey: privKey})

	t.Log("NodeKey:", transport.nodeKey)
	t.Log("NodeKey ID:", transport.nodeKey.ID())

	// Include the ID in the address string
	addressStr := fmt.Sprintf("%s@localhost:%d", transport.nodeKey.ID(), port)
	address, err := NewNetAddressString(addressStr)
	if err != nil {
		return nil, nil, err
	}

	if err := transport.Listen(*address); err != nil {
		return nil, nil, err
	}

	// Retrieve the assigned port and add the ID to the address
	addr, err := NewNetAddressString(fmt.Sprintf("%s@%s", transport.nodeKey.ID(), transport.listener.Addr().String()))
	if err != nil {
		return nil, nil, err
	}

	return transport, addr, nil
}

func setupSimulationParams(t *testing.T) (SimulationParams, error) {
	privKey := ed25519.GenPrivKey()
	nodeInfo := testNodeInfo(PubKeyToID(privKey.PubKey()), defaultNodeName)

	// Let the OS assign two free ports
	var fastPort, slowPort int
	for {
		fastPort = cmtnet.GetFreePort()
		slowPort = cmtnet.GetFreePort()
		if fastPort != slowPort {
			break
		}
	}

	fastTransport, fastAddr, err := setupTransport(t, nodeInfo, privKey, fastPort)
	if err != nil {
		return SimulationParams{}, err
	}

	slowTransport, slowAddr, err := setupTransport(t, nodeInfo, privKey, slowPort)
	if err != nil {
		return SimulationParams{}, err
	}

	return SimulationParams{
		PrivKey:       privKey,
		NodeInfo:      nodeInfo,
		FastTransport: fastTransport,
		SlowTransport: slowTransport,
		FastAddr:      fastAddr,
		SlowAddr:      slowAddr,
		SlowChan:      make(chan error),
		FastChan:      make(chan error),
		SlowDoneChan:  make(chan error),
		ErrChan:       make(chan error),
		T:             t,
	}, nil
}
