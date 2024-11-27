package p2p

import (
	"errors"
	"fmt"
	golog "log"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	p2p "github.com/cometbft/cometbft/api/cometbft/p2p/v1"
	"github.com/cometbft/cometbft/config"
	"github.com/cometbft/cometbft/crypto"
	"github.com/cometbft/cometbft/crypto/ed25519"
	"github.com/cometbft/cometbft/libs/bytes"
	"github.com/cometbft/cometbft/libs/log"
	na "github.com/cometbft/cometbft/p2p/netaddr"
	ni "github.com/cometbft/cometbft/p2p/nodeinfo"
	"github.com/cometbft/cometbft/p2p/nodekey"
	"github.com/cometbft/cometbft/p2p/transport"
	tcpconn "github.com/cometbft/cometbft/p2p/transport/tcp/conn"
)

const testCh = 0x01

func TestPeerBasic(t *testing.T) {
	rp := &remotePeer{PrivKey: ed25519.GenPrivKey(), Config: cfg}
	rp.Start()
	defer rp.Stop()

	p := createOutboundPeerAndPerformHandshake(t, rp.Addr(), cfg)

	err := p.Start()
	require.NoError(t, err)
	t.Cleanup(func() {
		if err := p.Stop(); err != nil {
			t.Error(err)
		}
	})

	assert.True(t, p.IsRunning())
	assert.True(t, p.IsOutbound())

	assert.False(t, p.IsPersistent())
	p.persistent = true
	assert.True(t, p.IsPersistent())

	assert.Equal(t, rp.Addr().DialString(), p.RemoteAddr().String())
	assert.Equal(t, rp.ID(), p.ID())
}

func TestPeerSend(t *testing.T) {
	config := cfg

	rp := &remotePeer{PrivKey: ed25519.GenPrivKey(), Config: config}
	rp.Start()
	defer rp.Stop()

	p := createOutboundPeerAndPerformHandshake(t, rp.Addr(), config)

	err := p.Start()
	require.NoError(t, err)
	t.Cleanup(func() {
		if err := p.Stop(); err != nil {
			t.Error(err)
		}
	})

	assert.Nil(t, p.Send(Envelope{ChannelID: testCh, Message: &p2p.Message{}}))
}

func createOutboundPeerAndPerformHandshake(
	t *testing.T,
	addr *na.NetAddr,
	config *config.P2PConfig,
) *peer {
	t.Helper()

	pc, err := testOutboundPeerConn(addr, config, false)
	require.NoError(t, err)

	// create dummy node info and perform handshake
	var (
		timeout     = 1 * time.Second
		ourNodeID   = nodekey.PubKeyToID(ed25519.GenPrivKey().PubKey())
		ourNodeInfo = testNodeInfo(ourNodeID, "host_peer")
	)
	peerNodeInfo, err := handshake(ourNodeInfo, pc, timeout)
	require.NoError(t, err)

	// create peer
	var (
		streamDescs = []transport.StreamDescriptor{
			tcpconn.StreamDescriptor{
				ID:           testCh,
				Priority:     1,
				MessageTypeI: &p2p.Message{},
			},
		}
		streamInfoByStreamID = map[byte]streamInfo{
			testCh: {
				reactor: NewTestReactor(streamDescs, true),
				msgType: &p2p.Message{},
			},
		}
	)
	p := newPeer(pc, peerNodeInfo, streamInfoByStreamID, func(_ Peer, _ any) {})
	p.SetLogger(log.TestingLogger().With("peer", addr))
	return p
}

func testDial(addr *na.NetAddr, cfg *config.P2PConfig) (transport.Connection, error) {
	if cfg.TestDialFail {
		return nil, errors.New("dial err (peerConfig.DialFail == true)")
	}
	conn, err := addr.DialTimeout(cfg.DialTimeout)
	if err != nil {
		return nil, err
	}
	return newMockConnection(conn), nil
}

// testOutboundPeerConn dials a remote peer and returns a peerConn.
// It ensures the dialed ID matches the connection ID.
func testOutboundPeerConn(addr *na.NetAddr, config *config.P2PConfig, persistent bool) (peerConn, error) {
	var pc peerConn

	conn, err := testDial(addr, config)
	if err != nil {
		return pc, fmt.Errorf("creating peer: %w", err)
	}

	pc, err = testPeerConn(conn, true, persistent, addr)
	if err != nil {
		_ = conn.Close(err.Error())
		return pc, err
	}

	if addr.ID != pc.ID() { // ensure dialed ID matches connection ID
		_ = conn.Close("dialed ID does not match connection ID")
		return pc, ErrSwitchAuthenticationFailure{addr, pc.ID()}
	}

	return pc, nil
}

type remotePeer struct {
	PrivKey  crypto.PrivKey
	Config   *config.P2PConfig
	addr     *na.NetAddr
	channels bytes.HexBytes
	listener net.Listener
}

func (rp *remotePeer) Addr() *na.NetAddr {
	return rp.addr
}

func (rp *remotePeer) ID() nodekey.ID {
	return nodekey.PubKeyToID(rp.PrivKey.PubKey())
}

func (rp *remotePeer) Start() {
	l, e := net.Listen("tcp", "127.0.0.1:0") // any available address
	if e != nil {
		golog.Fatalf("net.Listen tcp :0: %+v", e)
	}
	rp.listener = l

	rp.addr = na.New(nodekey.PubKeyToID(rp.PrivKey.PubKey()), l.Addr())

	rp.channels = []byte{testCh}

	go rp.accept()
}

func (rp *remotePeer) Stop() {
	rp.listener.Close()
}

func (rp *remotePeer) Dial(addr *na.NetAddr) (transport.Connection, error) {
	pc, err := testOutboundPeerConn(addr, rp.Config, false)
	if err != nil {
		return nil, err
	}

	_, err = handshake(rp.nodeInfo(), pc, time.Second)
	if err != nil {
		return nil, err
	}
	return pc, err
}

func (rp *remotePeer) accept() {
	conns := []peerConn{}

	for {
		netConn, err := rp.listener.Accept()
		if err != nil {
			golog.Printf("Failed to accept conn: %+v", err)
			for _, conn := range conns {
				_ = conn.Close(err.Error())
			}
			return
		}

		conn := newMockConnection(netConn)

		ni, err := handshake(rp.nodeInfo(), conn, time.Second)
		if err != nil {
			_ = conn.Close(err.Error())
			golog.Printf("Failed to perform handshake: %+v", err)
		}

		addr, _ := ni.NetAddr()
		pc, err := testInboundPeerConn(conn, addr)
		if err != nil {
			_ = conn.Close(err.Error())
			golog.Fatalf("Failed to create a peer: %+v", err)
		}

		conns = append(conns, pc)
	}
}

func (rp *remotePeer) nodeInfo() ni.NodeInfo {
	la := rp.listener.Addr().String()
	nodeInfo := testNodeInfo(rp.ID(), "remote_peer_"+la)
	nodeInfo.ListenAddr = la
	nodeInfo.Channels = rp.channels
	return nodeInfo
}

type mockStream struct {
	net.Conn
}

func (s mockStream) Read(b []byte) (n int, err error) {
	return s.Conn.Read(b)
}

func (s mockStream) Write(b []byte) (n int, err error) {
	return s.Conn.Write(b)
}

func (s mockStream) TryWrite(b []byte) (n int, err error) {
	return s.Conn.Write(b)
}

func (mockStream) Close() error {
	return nil
}
func (s mockStream) SetDeadline(t time.Time) error     { return s.Conn.SetReadDeadline(t) }
func (s mockStream) SetReadDeadline(t time.Time) error { return s.Conn.SetReadDeadline(t) }

type mockConnection struct {
	net.Conn
	connectedAt time.Time
}

func newMockConnection(c net.Conn) *mockConnection {
	return &mockConnection{
		Conn:        c,
		connectedAt: time.Now(),
	}
}

func (c mockConnection) OpenStream(byte, any) (transport.Stream, error) {
	return &mockStream{
		Conn: c.Conn,
	}, nil
}

func (c mockConnection) LocalAddr() net.Addr {
	return c.Conn.LocalAddr()
}

func (c mockConnection) RemoteAddr() net.Addr {
	return c.Conn.RemoteAddr()
}
func (c mockConnection) Close(string) error         { return c.Conn.Close() }
func (c mockConnection) FlushAndClose(string) error { return c.Conn.Close() }
func (mockConnection) ErrorCh() <-chan error        { return nil }

func (c mockConnection) ConnState() transport.ConnState {
	return transport.ConnState{
		ConnectedFor: time.Since(c.connectedAt),
	}
}
