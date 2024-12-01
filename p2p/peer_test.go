package p2p

import (
	"errors"
	"fmt"
	golog "log"
	"net"
	"testing"
	"time"

	"github.com/cosmos/gogoproto/proto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	p2p "github.com/cometbft/cometbft/api/cometbft/p2p/v1"
	"github.com/cometbft/cometbft/config"
	"github.com/cometbft/cometbft/crypto"
	"github.com/cometbft/cometbft/crypto/ed25519"
	"github.com/cometbft/cometbft/libs/bytes"
	"github.com/cometbft/cometbft/libs/log"
	ni "github.com/cometbft/cometbft/p2p/internal/nodeinfo"
	na "github.com/cometbft/cometbft/p2p/netaddr"
	"github.com/cometbft/cometbft/p2p/nodekey"
	tcpconn "github.com/cometbft/cometbft/p2p/transport/tcp/conn"
)

func TestPeerBasic(t *testing.T) {
	assert, require := assert.New(t), require.New(t)

	// simulate remote peer
	rp := &remotePeer{PrivKey: ed25519.GenPrivKey(), Config: cfg}
	rp.Start()
	t.Cleanup(rp.Stop)

	p, err := createOutboundPeerAndPerformHandshake(rp.Addr(), cfg, tcpconn.DefaultMConnConfig())
	require.NoError(err)

	err = p.Start()
	require.NoError(err)
	t.Cleanup(func() {
		if err := p.Stop(); err != nil {
			t.Error(err)
		}
	})

	assert.True(p.IsRunning())
	assert.True(p.IsOutbound())
	assert.False(p.IsPersistent())
	p.persistent = true
	assert.True(p.IsPersistent())
	assert.Equal(rp.Addr().DialString(), p.RemoteAddr().String())
	assert.Equal(rp.ID(), p.ID())
}

func TestPeerSend(t *testing.T) {
	assert, require := assert.New(t), require.New(t)

	config := cfg

	// simulate remote peer
	rp := &remotePeer{PrivKey: ed25519.GenPrivKey(), Config: config}
	rp.Start()
	t.Cleanup(rp.Stop)

	p, err := createOutboundPeerAndPerformHandshake(rp.Addr(), config, tcpconn.DefaultMConnConfig())
	require.NoError(err)

	err = p.Start()
	require.NoError(err)

	t.Cleanup(func() {
		if err := p.Stop(); err != nil {
			t.Error(err)
		}
	})

	assert.True(p.CanSend(testCh))
	assert.True(p.Send(Envelope{ChannelID: testCh, Message: &p2p.Message{}}))
}

func createOutboundPeerAndPerformHandshake(
	addr *na.NetAddr,
	config *config.P2PConfig,
	mConfig tcpconn.MConnConfig,
) (*peer, error) {
	// create outbound peer connection
	pc, err := testOutboundPeerConn(addr, config, false)
	if err != nil {
		return nil, err
	}

	// create dummy node info and perform handshake
	var (
		timeout     = 1 * time.Second
		ourNodeID   = nodekey.PubKeyToID(ed25519.GenPrivKey().PubKey())
		ourNodeInfo = testNodeInfo(ourNodeID, "host_peer")
	)
	peerNodeInfo, err := handshake(ourNodeInfo, pc.conn, timeout)
	if err != nil {
		return nil, err
	}

	// create peer
	var (
		streamDescs = []StreamDescriptor{
			&tcpconn.ChannelDescriptor{
				ID:           testCh,
				Priority:     1,
				MessageTypeI: &p2p.Message{},
			},
		}
		reactorsByCh  = map[byte]Reactor{testCh: NewTestReactor(streamDescs, true)}
		msgTypeByChID = map[byte]proto.Message{
			testCh: &p2p.Message{},
		}
	)
	p := newPeer(pc, mConfig, peerNodeInfo, reactorsByCh, msgTypeByChID, streamDescs, func(_ Peer, _ any) {})
	p.SetLogger(log.TestingLogger().With("peer", addr))
	return p, nil
}

func testDial(addr *na.NetAddr, cfg *config.P2PConfig) (net.Conn, error) {
	if cfg.TestDialFail {
		return nil, errors.New("dial err (peerConfig.DialFail == true)")
	}

	conn, err := addr.DialTimeout(cfg.DialTimeout)
	if err != nil {
		return nil, err
	}
	return conn, nil
}

func testOutboundPeerConn(
	addr *na.NetAddr,
	config *config.P2PConfig,
	persistent bool,
	// ourNodePrivKey crypto.PrivKey,
) (peerConn, error) {
	var pc peerConn
	conn, err := testDial(addr, config)
	if err != nil {
		return pc, fmt.Errorf("error creating peer: %w", err)
	}

	pc, err = testPeerConn(conn, config, true, persistent, addr)
	if err != nil {
		if cerr := conn.Close(); cerr != nil {
			return pc, fmt.Errorf("%v: %w", cerr.Error(), err)
		}
		return pc, err
	}

	// ensure dialed ID matches connection ID
	if addr.ID != pc.ID() {
		if cerr := conn.Close(); cerr != nil {
			return pc, fmt.Errorf("%v: %w", cerr.Error(), err)
		}
		return pc, ErrSwitchAuthenticationFailure{addr, pc.ID()}
	}

	return pc, nil
}

type remotePeer struct {
	PrivKey    crypto.PrivKey
	Config     *config.P2PConfig
	addr       *na.NetAddr
	channels   bytes.HexBytes
	listenAddr string
	listener   net.Listener
}

func (rp *remotePeer) Addr() *na.NetAddr {
	return rp.addr
}

func (rp *remotePeer) ID() nodekey.ID {
	return nodekey.PubKeyToID(rp.PrivKey.PubKey())
}

func (rp *remotePeer) Start() {
	if rp.listenAddr == "" {
		rp.listenAddr = "127.0.0.1:0"
	}

	l, e := net.Listen("tcp", rp.listenAddr) // any available address
	if e != nil {
		golog.Fatalf("net.Listen tcp :0: %+v", e)
	}
	rp.listener = l
	rp.addr = na.New(nodekey.PubKeyToID(rp.PrivKey.PubKey()), l.Addr())
	if rp.channels == nil {
		rp.channels = []byte{testCh}
	}
	go rp.accept()
}

func (rp *remotePeer) Stop() {
	rp.listener.Close()
}

func (rp *remotePeer) Dial(addr *na.NetAddr) (net.Conn, error) {
	pc, err := testOutboundPeerConn(addr, rp.Config, false)
	if err != nil {
		return nil, err
	}

	_, err = handshake(rp.nodeInfo(), pc.conn, time.Second)
	if err != nil {
		return nil, err
	}
	return pc.conn, err
}

func (rp *remotePeer) accept() {
	conns := []net.Conn{}

	for {
		conn, err := rp.listener.Accept()
		if err != nil {
			golog.Printf("Failed to accept conn: %+v", err)
			for _, conn := range conns {
				_ = conn.Close()
			}
			return
		}

		pc, err := testInboundPeerConn(conn, rp.Config)
		if err != nil {
			_ = conn.Close()
			golog.Fatalf("Failed to create a peer: %+v", err)
		}

		_, err = handshake(rp.nodeInfo(), pc.conn, time.Second)
		if err != nil {
			_ = pc.conn.Close()
			golog.Printf("Failed to perform handshake: %+v", err)
		}

		conns = append(conns, conn)
	}
}

func (rp *remotePeer) nodeInfo() ni.NodeInfo {
	la := rp.listener.Addr().String()
	nodeInfo := testNodeInfo(rp.ID(), "remote_peer_"+la)
	nodeInfo.ListenAddr = la
	nodeInfo.Channels = rp.channels
	return nodeInfo
}
