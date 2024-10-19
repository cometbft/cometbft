package tcp

import (
	"errors"
	"math/rand"
	"net"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/cometbft/cometbft/crypto/ed25519"
	"github.com/cometbft/cometbft/libs/log"
	"github.com/cometbft/cometbft/p2p/abstract"
	na "github.com/cometbft/cometbft/p2p/netaddr"
	"github.com/cometbft/cometbft/p2p/nodekey"
	"github.com/cometbft/cometbft/p2p/transport/tcp/conn"
)

// newMultiplexTransport returns a tcp connected multiplexed peer
// using the default MConnConfig. It's a convenience function used
// for testing.
func newMultiplexTransport(
	nodeKey nodekey.NodeKey,
) *MultiplexTransport {
	return NewMultiplexTransport(
		nodeKey, conn.DefaultMConnConfig(),
	)
}

func TestTransportMultiplex_ConnFilter(t *testing.T) {
	mt := newMultiplexTransport(
		nodekey.NodeKey{
			PrivKey: ed25519.GenPrivKey(),
		},
	)
	mt.SetLogger(log.TestingLogger())
	id := mt.nodeKey.ID()

	MultiplexTransportConnFilters(
		func(_ ConnSet, _ net.Conn, _ []net.IP) error { return nil },
		func(_ ConnSet, _ net.Conn, _ []net.IP) error { return nil },
		func(_ ConnSet, _ net.Conn, _ []net.IP) error {
			return errors.New("rejected")
		},
	)(mt)

	addr, err := na.NewFromString(na.IDAddrString(id, "127.0.0.1:0"))
	if err != nil {
		t.Fatal(err)
	}

	if err := mt.Listen(*addr); err != nil {
		t.Fatal(err)
	}

	errc := make(chan error)

	go func() {
		addr := na.New(id, mt.listener.Addr())

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

	_, _, err = mt.Accept()
	if e, ok := err.(ErrRejected); ok {
		if !e.IsFiltered() {
			t.Errorf("expected peer to be filtered, got %v", err)
		}
	} else {
		t.Errorf("expected ErrRejected, got %v", err)
	}
}

func TestTransportMultiplex_ConnFilterTimeout(t *testing.T) {
	mt := newMultiplexTransport(
		nodekey.NodeKey{
			PrivKey: ed25519.GenPrivKey(),
		},
	)
	mt.SetLogger(log.TestingLogger())
	id := mt.nodeKey.ID()

	MultiplexTransportFilterTimeout(5 * time.Millisecond)(mt)
	MultiplexTransportConnFilters(
		func(_ ConnSet, _ net.Conn, _ []net.IP) error {
			time.Sleep(1 * time.Second)
			return nil
		},
	)(mt)

	addr, err := na.NewFromString(na.IDAddrString(id, "127.0.0.1:0"))
	if err != nil {
		t.Fatal(err)
	}

	if err := mt.Listen(*addr); err != nil {
		t.Fatal(err)
	}

	errc := make(chan error)
	go func() {
		addr := na.New(id, mt.listener.Addr())

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

	_, _, err = mt.Accept()
	if _, ok := err.(ErrFilterTimeout); !ok {
		t.Errorf("expected ErrFilterTimeout, got %v", err)
	}
}

func TestTransportMultiplex_MaxIncomingConnections(t *testing.T) {
	pv := ed25519.GenPrivKey()
	id := nodekey.PubKeyToID(pv.PubKey())
	mt := newMultiplexTransport(
		nodekey.NodeKey{
			PrivKey: pv,
		},
	)

	MultiplexTransportMaxIncomingConnections(0)(mt)

	addr, err := na.NewFromString(na.IDAddrString(id, "127.0.0.1:0"))
	if err != nil {
		t.Fatal(err)
	}
	const maxIncomingConns = 2
	MultiplexTransportMaxIncomingConnections(maxIncomingConns)(mt)
	if err := mt.Listen(*addr); err != nil {
		t.Fatal(err)
	}

	laddr := na.New(mt.nodeKey.ID(), mt.listener.Addr())

	// Connect more peers than max
	for i := 0; i <= maxIncomingConns; i++ {
		errc := make(chan error)
		go testDialer(*laddr, errc)

		err = <-errc
		if i < maxIncomingConns {
			if err != nil {
				t.Errorf("dialer connection failed: %v", err)
			}
			_, _, err = mt.Accept()
			if err != nil {
				t.Errorf("connection failed: %v", err)
			}
		} else if err == nil || !strings.Contains(err.Error(), "i/o timeout") {
			// mt actually blocks forever on trying to accept a new peer into a full channel so
			// expect the dialer to encounter a timeout error. Calling mt.Accept will block until
			// mt is closed.
			t.Errorf("expected i/o timeout error, got %v", err)
		}
	}
}

func TestTransportMultiplex_AcceptMultiple(t *testing.T) {
	mt := testSetupMultiplexTransport(t)
	laddr := na.New(mt.nodeKey.ID(), mt.listener.Addr())

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

	conns := []abstract.Connection{}

	// Accept all connections.
	for i := 0; i < cap(errc); i++ {
		c, _, err := mt.Accept()
		if err != nil {
			t.Fatal(err)
		}

		conns = append(conns, c)
	}

	if have, want := len(conns), cap(errc); have != want {
		t.Errorf("have %v, want %v", have, want)
	}

	if err := mt.Close(); err != nil {
		t.Errorf("close errored: %v", err)
	}
}

func testDialer(dialAddr na.NetAddr, errc chan error) {
	var (
		pv     = ed25519.GenPrivKey()
		dialer = newMultiplexTransport(
			nodekey.NodeKey{
				PrivKey: pv,
			},
		)
	)
	dialer.SetLogger(log.TestingLogger())

	_, err := dialer.Dial(dialAddr)
	if err != nil {
		errc <- err
		return
	}

	// Signal that the connection was established.
	errc <- nil
}

func TestTransportMultiplexAcceptNonBlocking(t *testing.T) {
	mt := testSetupMultiplexTransport(t)

	var (
		fastNodePV = ed25519.GenPrivKey()
		errc       = make(chan error)
		fastc      = make(chan struct{})
		slowc      = make(chan struct{})
		slowdonec  = make(chan struct{})
	)

	// Simulate slow Peer.
	go func() {
		addr := na.New(mt.nodeKey.ID(), mt.listener.Addr())

		c, err := addr.Dial()
		if err != nil {
			errc <- err
			return
		}

		close(slowc)
		defer func() {
			close(slowdonec)
		}()

		// Make sure we switch to fast peer goroutine.
		runtime.Gosched()

		select {
		case <-fastc:
			// Fast peer connected.
		case <-time.After(200 * time.Millisecond):
			// We error if the fast peer didn't succeed.
			errc <- errors.New("fast peer timed out")
		}

		_, err = upgradeSecretConn(c, 200*time.Millisecond, ed25519.GenPrivKey())
		if err != nil {
			errc <- err
			return
		}
	}()

	// Simulate fast Peer.
	go func() {
		<-slowc

		dialer := newMultiplexTransport(
			nodekey.NodeKey{
				PrivKey: fastNodePV,
			},
		)
		dialer.SetLogger(log.TestingLogger())
		addr := na.New(mt.nodeKey.ID(), mt.listener.Addr())

		_, err := dialer.Dial(*addr)
		if err != nil {
			errc <- err
			return
		}

		close(fastc)
		<-slowdonec
		close(errc)
	}()

	if err := <-errc; err != nil {
		t.Logf("connection failed: %v", err)
	}

	_, _, err := mt.Accept()
	if err != nil {
		t.Fatal(err)
	}
}

func TestTransportMultiplexDialRejectWrongID(t *testing.T) {
	mt := testSetupMultiplexTransport(t)

	var (
		pv     = ed25519.GenPrivKey()
		dialer = newMultiplexTransport(
			nodekey.NodeKey{
				PrivKey: pv,
			},
		)
	)
	dialer.SetLogger(log.TestingLogger())

	wrongID := nodekey.PubKeyToID(ed25519.GenPrivKey().PubKey())
	addr := na.New(wrongID, mt.listener.Addr())

	_, err := dialer.Dial(*addr)
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

// create listener.
func testSetupMultiplexTransport(t *testing.T) *MultiplexTransport {
	t.Helper()

	var (
		pv = ed25519.GenPrivKey()
		id = nodekey.PubKeyToID(pv.PubKey())
		mt = newMultiplexTransport(
			nodekey.NodeKey{
				PrivKey: pv,
			},
		)
	)
	mt.SetLogger(log.TestingLogger())

	addr, err := na.NewFromString(na.IDAddrString(id, "127.0.0.1:0"))
	if err != nil {
		t.Fatal(err)
	}

	if err := mt.Listen(*addr); err != nil {
		t.Fatal(err)
	}

	// give the listener some time to get ready
	time.Sleep(20 * time.Millisecond)

	return mt
}

type testTransportAddr struct{}

func (*testTransportAddr) Network() string { return "tcp" }
func (*testTransportAddr) String() string  { return "test.local:1234" }

type testTransportConn struct{}

func (*testTransportConn) Close() error {
	return errors.New("close() not implemented")
}

func (*testTransportConn) LocalAddr() net.Addr {
	return &testTransportAddr{}
}

func (*testTransportConn) RemoteAddr() net.Addr {
	return &testTransportAddr{}
}

func (*testTransportConn) Read(_ []byte) (int, error) {
	return -1, errors.New("read() not implemented")
}

func (*testTransportConn) SetDeadline(_ time.Time) error {
	return errors.New("setDeadline() not implemented")
}

func (*testTransportConn) SetReadDeadline(_ time.Time) error {
	return errors.New("setReadDeadline() not implemented")
}

func (*testTransportConn) SetWriteDeadline(_ time.Time) error {
	return errors.New("setWriteDeadline() not implemented")
}

func (*testTransportConn) Write(_ []byte) (int, error) {
	return -1, errors.New("write() not implemented")
}
