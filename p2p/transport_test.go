package p2p

import (
	"fmt"
	"math/rand"
	"net"
	"reflect"

	"errors"
	"strings"
	"sync"
	"testing"
	"time"

	tmp2p "github.com/cometbft/cometbft/api/cometbft/p2p/v1"
	"github.com/cometbft/cometbft/crypto/ed25519"
	"github.com/cometbft/cometbft/internal/protoio"
	"github.com/cometbft/cometbft/p2p/conn"
	"github.com/stretchr/testify/require"

	"github.com/sirupsen/logrus"
)

var defaultNodeName = "host_peer"

func emptyNodeInfo() NodeInfo {
	return DefaultNodeInfo{}
}

// newMultiplexTransport returns a tcp connected multiplexed peer
// using the default MConnConfig. It's a convenience function used
// for testing.
func newMultiplexTransport(
	nodeInfo NodeInfo,
	nodeKey NodeKey,
) *MultiplexTransport {
	return NewMultiplexTransport(
		nodeInfo, nodeKey, conn.DefaultMConnConfig(),
	)
}

func TestTransportMultiplexConnFilter(t *testing.T) {
	t.Log("Creating new multiplex transport")
	mt := newMultiplexTransport(
		emptyNodeInfo(),
		NodeKey{
			PrivKey: ed25519.GenPrivKey(),
		},
	)
	id := mt.nodeKey.ID()

	t.Log("Setting up connection filters")
	MultiplexTransportConnFilters(
		func(_ ConnSet, _ net.Conn, _ []net.IP) error { return nil },
		func(_ ConnSet, _ net.Conn, _ []net.IP) error { return nil },
		func(_ ConnSet, _ net.Conn, _ []net.IP) error {
			return fmt.Errorf("rejected")
		},
	)(mt)

	t.Log("Creating new network address")
	addr, err := NewNetAddressString(IDAddressString(id, "127.0.0.1:0"))
	if err != nil {
		t.Fatalf("Failed to create network address: %v", err)
	}

	t.Log("Starting listener")
	if err := mt.Listen(*addr); err != nil {
		t.Fatalf("Failed to start listener: %v", err)
	}

	errc := make(chan error)

	go func() {
		t.Log("Dialing")
		addr := NewNetAddress(id, mt.listener.Addr())

		_, err := addr.Dial()
		if err != nil {
			errc <- fmt.Errorf("Failed to dial: %v", err)
			return
		}

		close(errc)
	}()

	t.Log("Waiting for dial to complete")
	if err := <-errc; err != nil {
		t.Errorf("Connection failed: %v", err)
	}

	t.Log("Accepting connection")
	_, err = mt.Accept(peerConfig{})
	if e, ok := err.(ErrRejected); ok {
		if !e.IsFiltered() {
			t.Errorf("Expected peer to be filtered, got %v", err)
		}
	} else {
		t.Errorf("Expected ErrRejected, got %v", err)
	}
}

func TestTransportMultiplexConnFilterTimeout(t *testing.T) {
	mt := newMultiplexTransport(
		emptyNodeInfo(),
		NodeKey{
			PrivKey: ed25519.GenPrivKey(),
		},
	)
	id := mt.nodeKey.ID()

	MultiplexTransportFilterTimeout(100 * time.Millisecond)(mt)
	MultiplexTransportConnFilters(
		func(_ ConnSet, _ net.Conn, _ []net.IP) error {
			time.Sleep(1 * time.Second)
			return nil
		},
	)(mt)

	addr, err := NewNetAddressString(IDAddressString(id, "127.0.0.1:0"))
	if err != nil {
		t.Fatal(err)
	}

	if err := mt.Listen(*addr); err != nil {
		t.Fatal(err)
	}

	errc := make(chan error)
	go func() {
		addr := NewNetAddress(id, mt.listener.Addr())

		_, err := addr.Dial()
		if err != nil {
			errc <- err
			return
		}

		close(errc)
	}()

	if err := <-errc; err != nil {
		t.Errorf("connection failed: %v", err)
	}

	_, err = mt.Accept(peerConfig{})
	if _, ok := err.(ErrFilterTimeout); !ok {
		t.Errorf("expected ErrFilterTimeout, got %v", err)
	}
}

func TestTransportMultiplexMaxIncomingConnections(t *testing.T) {
	pv := ed25519.GenPrivKey()
	id := PubKeyToID(pv.PubKey())
	mt := newMultiplexTransport(
		testNodeInfo(
			id, "transport",
		),
		NodeKey{
			PrivKey: pv,
		},
	)

	t.Log("Setting maximum incoming connections to 0")
	MultiplexTransportMaxIncomingConnections(0)(mt)

	addr, err := NewNetAddressString(IDAddressString(id, "127.0.0.1:0"))
	if err != nil {
		t.Fatalf("Failed to create network address: %v", err)
	}

	const maxIncomingConns = 3
	t.Logf("Setting maximum incoming connections to %d", maxIncomingConns)
	MultiplexTransportMaxIncomingConnections(maxIncomingConns)(mt)

	t.Log("Starting listener")
	if err := mt.Listen(*addr); err != nil {
		t.Fatalf("Failed to start listener: %v", err)
	}

	laddr := NewNetAddress(mt.nodeKey.ID(), mt.listener.Addr())

	// Connect more peers than max
	for i := 0; i <= maxIncomingConns; i++ {
		t.Logf("Dialing connection %d", i+1)
		errc := make(chan error)
		go testDialer(*laddr, errc)

		err = <-errc
		if i < maxIncomingConns {
			if err != nil {
				t.Errorf("Dialer connection %d failed: %v", i+1, err)
			}
			_, err = mt.Accept(peerConfig{})
			if err != nil {
				t.Errorf("Accepting connection %d failed: %v", i+1, err)
			}
		} else {
			if err == nil || !strings.Contains(err.Error(), "i/o timeout") {
				// mt actually blocks forever on trying to accept a new peer into a full channel so
				// expect the dialer to encounter a timeout error. Calling mt.Accept will block until
				// mt is closed.
				t.Errorf("Expected i/o timeout error for connection %d, got %v", i+1, err)
			}
		}
	}
}

func TestTransportMultiplexAcceptMultiple(t *testing.T) {
	mt := testSetupMultiplexTransport(t)
	laddr := NewNetAddress(mt.nodeKey.ID(), mt.listener.Addr())

	var (
		seed     = rand.New(rand.NewSource(time.Now().UnixNano()))
		nDialers = seed.Intn(64) + 64
		errc     = make(chan error, nDialers)
	)

	// Setup dialers.
	for i := 0; i < nDialers; i++ {
		go testDialer(*laddr, errc)
	}

	// Catch connection errors.
	for i := 0; i < nDialers; i++ {
		if err := <-errc; err != nil {
			t.Fatal(err)
		}
	}

	ps := []Peer{}

	// Accept all peers.
	for i := 0; i < cap(errc); i++ {
		p, err := mt.Accept(peerConfig{})
		if err != nil {
			t.Fatal(err)
		}

		if err := p.Start(); err != nil {
			t.Fatal(err)
		}

		ps = append(ps, p)
	}

	if have, want := len(ps), cap(errc); have != want {
		t.Errorf("have %v, want %v", have, want)
	}

	// Stop all peers.
	for _, p := range ps {
		if err := p.Stop(); err != nil {
			t.Fatal(err)
		}
	}

	if err := mt.Close(); err != nil {
		t.Errorf("close errored: %v", err)
	}
}

func testDialer(dialAddr NetAddress, errc chan error) {
	var (
		pv     = ed25519.GenPrivKey()
		dialer = newMultiplexTransport(
			testNodeInfo(PubKeyToID(pv.PubKey()), defaultNodeName),
			NodeKey{
				PrivKey: pv,
			},
		)
	)

	_, err := dialer.Dial(dialAddr, peerConfig{})
	if err != nil {
		errc <- err
		return
	}

	// Signal that the connection was established.
	errc <- nil
}

func TestTransportMultiplexAcceptNonBlocking(t *testing.T) {
	var passCount int
	var closedConnCount int
	totalIterations := 100
	failureModes := make(map[string]int)

	log := logrus.New()
	log.SetFormatter(&logrus.TextFormatter{
		DisableColors: false,
		FullTimestamp: true,
	})

	for i := 0; i < totalIterations; i++ {
		log.Infof("Test iteration %d started", i)

		func() {
			// Setup a multiplex transport
			mt := testSetupMultiplexTransport(t)
			log.Info("Multiplex transport set up")

			fastNodePV := ed25519.GenPrivKey()
			fastNodeInfo := testNodeInfo(PubKeyToID(fastNodePV.PubKey()), "fastnode")
			errc := make(chan error, 2) // Buffer to prevent goroutine leaks
			fastc := make(chan struct{})
			slowc := make(chan struct{})

			slowPeerPV := ed25519.GenPrivKey()
			defer mt.Close() // Ensure resources are cleaned up

			var wg sync.WaitGroup
			wg.Add(2)

			// Simulate slow Peer.
			go func() {
				time.Sleep(100 * time.Millisecond) // Delay slow peer
				log.Info("Slow peer starting")
				addr := NewNetAddress(mt.nodeKey.ID(), mt.listener.Addr())
				dialer := newMultiplexTransport(
					testNodeInfo(PubKeyToID(slowPeerPV.PubKey()), "slowpeer"),
					NodeKey{
						PrivKey: slowPeerPV,
					},
				)
				_, err := dialer.Dial(*addr, peerConfig{})
				if err != nil {
					errc <- errors.New(fmt.Sprintf("slow peer failed to dial: %v", err))
					return
				}
				log.Info("Peer connected")

				// Signal that the connection was established.
				close(slowc)
				wg.Done()
			}()

			// Simulate fast Peer.
			go func() {
				log.Info("Fast peer starting")
				addr := NewNetAddress(mt.nodeKey.ID(), mt.listener.Addr())
				dialer := newMultiplexTransport(
					fastNodeInfo,
					NodeKey{
						PrivKey: fastNodePV,
					},
				)
				_, err := dialer.Dial(*addr, peerConfig{})
				if err != nil {
					errc <- errors.New(fmt.Sprintf("fast peer failed to dial: %v", err))
					return
				}

				// Wait until the slow peer has connected.
				<-slowc

				log.Info("Fast peer finished")
				close(fastc)
				wg.Done()
			}()

			// Wait for both peers to finish
			wg.Wait()

			// Check for errors
			close(errc)
			for err := range errc {
				if err != nil {
					t.Errorf("connection failed: %v", err)
					failureModes[err.Error()]++
					if strings.Contains(err.Error(), "use of closed network connection") {
						closedConnCount++
					}
				}
			}

			p, err := mt.Accept(peerConfig{})
			if err != nil {
				require.NoError(t, err, "failed to accept peer")
			}

			if have, want := p.NodeInfo().ID(), fastNodeInfo.ID(); have != want {
				t.Errorf("have %v, want %v", have, want)
				t.Logf("Mismatched Node ID: have %v, want %v", have, want)
				t.Logf("Fast peer NodeInfo: %v", fastNodeInfo)
				t.Logf("Accepted peer NodeInfo: %v", p.NodeInfo())
				failureModes["Mismatched Node ID"]++
			} else {
				log.Info("Fast peer's NodeInfo correctly received")
				passCount++
			}
			log.Infof("Test iteration %d ended", i)
		}()

	}

	log.Infof("Pass rate: %.2f%%", float64(passCount)/float64(totalIterations)*100)
	log.Infof("Closed connection failures: %d", closedConnCount)
	for mode, count := range failureModes {
		log.Infof("Failure mode %s: %.2f%%", mode, float64(count)/float64(totalIterations)*100)
	}

}

func TestTransportMultiplexValidateNodeInfo(t *testing.T) {
	mt := testSetupMultiplexTransport(t)

	errc := make(chan error)

	go func() {
		var (
			pv     = ed25519.GenPrivKey()
			dialer = newMultiplexTransport(
				testNodeInfo(PubKeyToID(pv.PubKey()), ""), // Should not be empty
				NodeKey{
					PrivKey: pv,
				},
			)
		)

		addr := NewNetAddress(mt.nodeKey.ID(), mt.listener.Addr())

		_, err := dialer.Dial(*addr, peerConfig{})
		if err != nil {
			errc <- err
			return
		}

		close(errc)
	}()

	if err := <-errc; err != nil {
		t.Errorf("connection failed: %v", err)
	}

	_, err := mt.Accept(peerConfig{})
	if e, ok := err.(ErrRejected); ok {
		if !e.IsNodeInfoInvalid() {
			t.Errorf("expected NodeInfo to be invalid, got %v", err)
		}
	} else {
		t.Errorf("expected ErrRejected, got %v", err)
	}
}

func TestTransportMultiplexRejectMissmatchID(t *testing.T) {
	t.Log("Setting up multiplex transport")
	mt := testSetupMultiplexTransport(t)

	errc := make(chan error)

	go func() {
		t.Log("Creating dialer with mismatched ID")
		dialer := newMultiplexTransport(
			testNodeInfo(
				PubKeyToID(ed25519.GenPrivKey().PubKey()), "dialer",
			),
			NodeKey{
				PrivKey: ed25519.GenPrivKey(),
			},
		)
		addr := NewNetAddress(mt.nodeKey.ID(), mt.listener.Addr())

		t.Log("Dialing")
		_, err := dialer.Dial(*addr, peerConfig{})
		if err != nil {
			t.Log("Dialing failed")
			errc <- err
			return
		}

		t.Log("Dialing succeeded")
		close(errc)
	}()

	t.Log("Waiting for dialing result")
	if err := <-errc; err != nil {
		t.Errorf("connection failed: %v", err)
	}

	t.Log("Accepting connection")
	_, err := mt.Accept(peerConfig{})
	if e, ok := err.(ErrRejected); ok {
		if !e.IsAuthFailure() {
			t.Errorf("expected auth failure, got %v", e)
		} else {
			t.Log("Auth failure as expected")
		}
	} else {
		t.Errorf("expected ErrRejected, got %v", err)
	}
}

func TestTransportMultiplexDialRejectWrongID(t *testing.T) {
	mt := testSetupMultiplexTransport(t)

	var (
		pv     = ed25519.GenPrivKey()
		dialer = newMultiplexTransport(
			testNodeInfo(PubKeyToID(pv.PubKey()), ""), // Should not be empty
			NodeKey{
				PrivKey: pv,
			},
		)
	)

	wrongID := PubKeyToID(ed25519.GenPrivKey().PubKey())
	addr := NewNetAddress(wrongID, mt.listener.Addr())

	_, err := dialer.Dial(*addr, peerConfig{})
	if err != nil {
		t.Logf("connection failed: %v", err)
		if e, ok := err.(ErrRejected); ok {
			if !e.IsAuthFailure() {
				t.Errorf("expected auth failure, got %v", e)
			}
		} else {
			t.Errorf("expected ErrRejected, got %v", err)
		}
	}
}

func TestTransportMultiplexRejectIncompatible(t *testing.T) {
	mt := testSetupMultiplexTransport(t)

	errc := make(chan error)

	go func() {
		var (
			pv     = ed25519.GenPrivKey()
			dialer = newMultiplexTransport(
				testNodeInfoWithNetwork(PubKeyToID(pv.PubKey()), "dialer", "incompatible-network"),
				NodeKey{
					PrivKey: pv,
				},
			)
		)
		addr := NewNetAddress(mt.nodeKey.ID(), mt.listener.Addr())

		_, err := dialer.Dial(*addr, peerConfig{})
		if err != nil {
			errc <- err
			return
		}

		close(errc)
	}()

	_, err := mt.Accept(peerConfig{})
	if e, ok := err.(ErrRejected); ok {
		if !e.IsIncompatible() {
			t.Errorf("expected to reject incompatible, got %v", e)
		}
	} else {
		t.Errorf("expected ErrRejected, got %v", err)
	}
}

func TestTransportMultiplexRejectSelf(t *testing.T) {
	mt := testSetupMultiplexTransport(t)

	errc := make(chan error)

	go func() {
		addr := NewNetAddress(mt.nodeKey.ID(), mt.listener.Addr())

		_, err := mt.Dial(*addr, peerConfig{})
		if err != nil {
			errc <- err
			return
		}

		close(errc)
	}()

	if err := <-errc; err != nil {
		if e, ok := err.(ErrRejected); ok {
			if !e.IsSelf() {
				t.Errorf("expected to reject self, got: %v", e)
			}
		} else {
			t.Errorf("expected ErrRejected, got %v", err)
		}
	} else {
		t.Errorf("expected connection failure")
	}

	_, err := mt.Accept(peerConfig{})
	if err, ok := err.(ErrRejected); ok {
		if !err.IsSelf() {
			t.Errorf("expected to reject self, got: %v", err)
		}
	} else {
		t.Errorf("expected ErrRejected, got %v", nil)
	}
}

func TestTransportConnDuplicateIPFilter(t *testing.T) {
	filter := ConnDuplicateIPFilter()

	if err := filter(nil, &testTransportConn{}, nil); err != nil {
		t.Fatal(err)
	}

	var (
		c  = &testTransportConn{}
		cs = NewConnSet()
	)

	cs.Set(c, []net.IP{
		{10, 0, 10, 1},
		{10, 0, 10, 2},
		{10, 0, 10, 3},
	})

	if err := filter(cs, c, []net.IP{
		{10, 0, 10, 2},
	}); err == nil {
		t.Errorf("expected Peer to be rejected as duplicate")
	}
}

func TestTransportHandshake(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}

	var (
		peerPV       = ed25519.GenPrivKey()
		peerNodeInfo = testNodeInfo(PubKeyToID(peerPV.PubKey()), defaultNodeName)
	)

	go func() {
		c, err := net.Dial(ln.Addr().Network(), ln.Addr().String())
		if err != nil {
			t.Error(err)
			return
		}

		go func(c net.Conn) {
			_, err := protoio.NewDelimitedWriter(c).WriteMsg(peerNodeInfo.(DefaultNodeInfo).ToProto())
			if err != nil {
				t.Error(err)
			}
		}(c)
		go func(c net.Conn) {
			// ni   DefaultNodeInfo
			var pbni tmp2p.DefaultNodeInfo

			protoReader := protoio.NewDelimitedReader(c, MaxNodeInfoSize())
			_, err := protoReader.ReadMsg(&pbni)
			if err != nil {
				t.Error(err)
			}

			_, err = DefaultNodeInfoFromToProto(&pbni)
			if err != nil {
				t.Error(err)
			}
		}(c)
	}()

	c, err := ln.Accept()
	if err != nil {
		t.Fatal(err)
	}

	ni, err := handshake(c, 50*time.Millisecond, emptyNodeInfo())
	if err != nil {
		t.Fatal(err)
	}

	if have, want := ni, peerNodeInfo; !reflect.DeepEqual(have, want) {
		t.Errorf("have %v, want %v", have, want)
	}
}

func TestTransportAddChannel(t *testing.T) {
	mt := newMultiplexTransport(
		emptyNodeInfo(),
		NodeKey{
			PrivKey: ed25519.GenPrivKey(),
		},
	)
	testChannel := byte(0x01)

	mt.AddChannel(testChannel)
	if !mt.nodeInfo.(DefaultNodeInfo).HasChannel(testChannel) {
		t.Errorf("missing added channel %v. Got %v", testChannel, mt.nodeInfo.(DefaultNodeInfo).Channels)
	}
}

// create listener
func testSetupMultiplexTransport(t *testing.T) *MultiplexTransport {
	var (
		pv = ed25519.GenPrivKey()
		id = PubKeyToID(pv.PubKey())
		mt = newMultiplexTransport(
			testNodeInfo(
				id, "transport",
			),
			NodeKey{
				PrivKey: pv,
			},
		)
	)

	addr, err := NewNetAddressString(IDAddressString(id, "127.0.0.1:0"))
	if err != nil {
		t.Fatal(err)
	}

	if err := mt.Listen(*addr); err != nil {
		t.Fatal(err)
	}

	// give the listener some time to get ready
	time.Sleep(50 * time.Millisecond)

	return mt
}

type testTransportAddr struct{}

func (a *testTransportAddr) Network() string { return "tcp" }
func (a *testTransportAddr) String() string  { return "test.local:1234" }

type testTransportConn struct{}

func (c *testTransportConn) Close() error {
	return fmt.Errorf("close() not implemented")
}

func (c *testTransportConn) LocalAddr() net.Addr {
	return &testTransportAddr{}
}

func (c *testTransportConn) RemoteAddr() net.Addr {
	return &testTransportAddr{}
}

func (c *testTransportConn) Read(_ []byte) (int, error) {
	return -1, fmt.Errorf("read() not implemented")
}

func (c *testTransportConn) SetDeadline(_ time.Time) error {
	return fmt.Errorf("setDeadline() not implemented")
}

func (c *testTransportConn) SetReadDeadline(_ time.Time) error {
	return fmt.Errorf("setReadDeadline() not implemented")
}

func (c *testTransportConn) SetWriteDeadline(_ time.Time) error {
	return fmt.Errorf("setWriteDeadline() not implemented")
}

func (c *testTransportConn) Write(_ []byte) (int, error) {
	return -1, fmt.Errorf("write() not implemented")
}
