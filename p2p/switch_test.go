package p2p

import (
	stdbytes "bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strconv"
	"sync/atomic"
	"testing"
	"time"

	"github.com/cosmos/gogoproto/proto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	p2pproto "github.com/cometbft/cometbft/api/cometbft/p2p/v1"
	"github.com/cometbft/cometbft/config"
	"github.com/cometbft/cometbft/crypto"
	"github.com/cometbft/cometbft/crypto/ed25519"
	"github.com/cometbft/cometbft/libs/log"
	cmtsync "github.com/cometbft/cometbft/libs/sync"
	na "github.com/cometbft/cometbft/p2p/netaddr"
	ni "github.com/cometbft/cometbft/p2p/nodeinfo"
	"github.com/cometbft/cometbft/p2p/nodekey"
	"github.com/cometbft/cometbft/p2p/transport"
	"github.com/cometbft/cometbft/p2p/transport/tcp"
	tcpconn "github.com/cometbft/cometbft/p2p/transport/tcp/conn"
)

var cfg *config.P2PConfig

func init() {
	cfg = config.DefaultP2PConfig()
	cfg.PexReactor = true
	cfg.AllowDuplicateIP = true
}

type PeerMessage struct {
	Contents proto.Message
	Counter  int
}

type TestReactor struct {
	BaseReactor

	mtx               cmtsync.Mutex
	streamDescriptors []transport.StreamDescriptor
	logMessages       bool
	msgsCounter       int
	msgsReceived      map[byte][]PeerMessage
}

func NewTestReactor(descs []transport.StreamDescriptor, logMessages bool) *TestReactor {
	tr := &TestReactor{
		streamDescriptors: descs,
		logMessages:       logMessages,
		msgsReceived:      make(map[byte][]PeerMessage),
	}
	tr.BaseReactor = *NewBaseReactor("TestReactor", tr)
	tr.SetLogger(log.TestingLogger())
	return tr
}

func (tr *TestReactor) StreamDescriptors() []transport.StreamDescriptor {
	return tr.streamDescriptors
}

func (*TestReactor) AddPeer(Peer) {}

func (*TestReactor) RemovePeer(Peer, any) {}

func (tr *TestReactor) Receive(e Envelope) {
	if tr.logMessages {
		tr.mtx.Lock()
		tr.msgsReceived[e.ChannelID] = append(tr.msgsReceived[e.ChannelID], PeerMessage{Contents: e.Message, Counter: tr.msgsCounter})
		tr.msgsCounter++
		tr.mtx.Unlock()
	}
}

func (tr *TestReactor) getMsgs(chID byte) []PeerMessage {
	tr.mtx.Lock()
	defer tr.mtx.Unlock()
	return tr.msgsReceived[chID]
}

// -----------------------------------------------------------------------------

// convenience method for creating two switches connected to each other.
// XXX: note this uses net.Pipe and not a proper TCP conn.
func MakeSwitchPair(initSwitch func(int, *Switch) *Switch) (*Switch, *Switch) {
	// Create two switches that will be interconnected.
	switches := MakeConnectedSwitches(cfg, 2, initSwitch, Connect2Switches)
	return switches[0], switches[1]
}

func initSwitchFunc(_ int, sw *Switch) *Switch {
	sw.SetAddrBook(&AddrBookMock{
		Addrs:    make(map[string]struct{}),
		OurAddrs: make(map[string]struct{}),
	})

	// Make two reactors of two channels each
	sw.AddReactor("foo", NewTestReactor([]transport.StreamDescriptor{
		tcpconn.StreamDescriptor{
			ID:           byte(0x00),
			Priority:     1,
			MessageTypeI: &p2pproto.Message{},
		},
		tcpconn.StreamDescriptor{
			ID:           byte(0x01),
			Priority:     2,
			MessageTypeI: &p2pproto.Message{},
		},
	}, true))
	sw.AddReactor("bar", NewTestReactor([]transport.StreamDescriptor{
		tcpconn.StreamDescriptor{
			ID:           byte(0x02),
			Priority:     3,
			MessageTypeI: &p2pproto.Message{},
		},
		tcpconn.StreamDescriptor{
			ID:           byte(0x03),
			Priority:     4,
			MessageTypeI: &p2pproto.Message{},
		},
	}, true))

	return sw
}

func TestSwitches(t *testing.T) {
	s1, s2 := MakeSwitchPair(initSwitchFunc)
	t.Cleanup(func() {
		if err := s2.Stop(); err != nil {
			t.Error(err)
		}
		if err := s1.Stop(); err != nil {
			t.Error(err)
		}
	})

	if s1.Peers().Size() != 1 {
		t.Errorf("expected exactly 1 peer in s1, got %v", s1.Peers().Size())
	}
	if s2.Peers().Size() != 1 {
		t.Errorf("expected exactly 1 peer in s2, got %v", s2.Peers().Size())
	}

	// Lets send some messages
	ch0Msg := &p2pproto.PexAddrs{
		Addrs: []p2pproto.NetAddress{
			{
				ID: "0",
			},
		},
	}
	ch1Msg := &p2pproto.PexAddrs{
		Addrs: []p2pproto.NetAddress{
			{
				ID: "1",
			},
		},
	}
	ch2Msg := &p2pproto.PexAddrs{
		Addrs: []p2pproto.NetAddress{
			{
				ID: "2",
			},
		},
	}
	// Test broadcast and TryBroadcast on different channels in parallel.
	// We have no channel capacity concerns, as each broadcast is on a distinct channel
	s1.Broadcast(Envelope{ChannelID: byte(0x00), Message: ch0Msg})
	s1.Broadcast(Envelope{ChannelID: byte(0x01), Message: ch1Msg})
	s1.TryBroadcast(Envelope{ChannelID: byte(0x02), Message: ch2Msg})
	assertMsgReceivedWithTimeout(t,
		ch0Msg,
		byte(0x00),
		s2.Reactor("foo").(*TestReactor), 200*time.Millisecond, 5*time.Second)
	assertMsgReceivedWithTimeout(t,
		ch1Msg,
		byte(0x01),
		s2.Reactor("foo").(*TestReactor), 200*time.Millisecond, 5*time.Second)
	assertMsgReceivedWithTimeout(t,
		ch2Msg,
		byte(0x02),
		s2.Reactor("bar").(*TestReactor), 200*time.Millisecond, 5*time.Second)
}

func assertMsgReceivedWithTimeout(
	t *testing.T,
	msg proto.Message,
	channel byte,
	reactor *TestReactor,
	checkPeriod,
	timeout time.Duration,
) {
	t.Helper()

	ticker := time.NewTicker(checkPeriod)
	defer ticker.Stop()

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	for {
		select {
		case <-ticker.C:
			msgs := reactor.getMsgs(channel)
			if len(msgs) != 0 {
				got, err := proto.Marshal(msgs[0].Contents)
				require.NoError(t, err)
				wanted, err := proto.Marshal(msg)
				require.NoError(t, err)
				if !stdbytes.Equal(got, wanted) {
					t.Fatalf("Unexpected message bytes. Wanted: %v, Got: %v", msg, msgs[0].Contents)
				}
				return
			}
		case <-ctx.Done():
			t.Fatalf("Expected to have received 1 message in channel #%v, but got 0", channel)
		}
	}
}

func TestSwitchFiltersOutItself(t *testing.T) {
	s1 := MakeSwitch(cfg, 1, initSwitchFunc)
	err := s1.Start()
	require.NoError(t, err)
	defer s1.Stop() //nolint:errcheck

	// simulate s1 having a public IP by creating a remote peer with the same ID
	rp := newRemoteTCPPeerWithPrivKey(s1.nodeKey.PrivKey)
	rp.Start()

	// addr should be rejected in addPeer based on the same ID
	err = s1.DialPeerWithAddress(rp.Addr())
	require.Error(t, err)
	if errR, ok := err.(ErrRejected); ok {
		if !errR.IsSelf() {
			t.Fatalf("expected self to be rejected, but got %v", errR)
		}
	} else {
		t.Fatalf("expected ErrRejected, but got %v", err)
	}

	assert.True(t, s1.addrBook.OurAddress(rp.Addr()))
	assert.False(t, s1.addrBook.HasAddress(rp.Addr()))

	rp.Stop()

	assertNoPeersAfterTimeout(t, s1, 100*time.Millisecond)
}

func TestSwitchPeerFilter(t *testing.T) {
	var (
		filters = []PeerFilterFunc{
			func(_ IPeerSet, _ Peer) error { return nil },
			func(_ IPeerSet, _ Peer) error { return errors.New("denied") },
			func(_ IPeerSet, _ Peer) error { return nil },
		}
		sw = MakeSwitch(
			cfg,
			1,
			initSwitchFunc,
			SwitchPeerFilters(filters...),
		)
	)
	err := sw.Start()
	require.NoError(t, err)
	t.Cleanup(func() {
		if err := sw.Stop(); err != nil {
			t.Error(err)
		}
	})

	// simulate remote peer
	rp := newRemoteTCPPeer()
	rp.Start()
	t.Cleanup(rp.Stop)

	conn, err := sw.transport.Dial(*rp.Addr())
	if err != nil {
		t.Fatal(err)
	}

	p := wrapPeer(conn,
		rp.nodeInfo(),
		peerConfig{
			onPeerError:          sw.StopPeerForError,
			isPersistent:         sw.IsPeerPersistent,
			streamInfoByStreamID: sw.streamInfoByStreamID,
			metrics:              sw.metrics,
			outbound:             true,
		},
		rp.Addr())

	err = sw.addPeer(p)
	if err, ok := err.(ErrRejected); ok {
		if !err.IsFiltered() {
			t.Errorf("expected peer to be filtered")
		}
	} else {
		t.Errorf("expected ErrRejected")
	}
}

func TestSwitchPeerFilterTimeout(t *testing.T) {
	var (
		filters = []PeerFilterFunc{
			func(_ IPeerSet, _ Peer) error {
				time.Sleep(10 * time.Millisecond)
				return nil
			},
		}
		sw = MakeSwitch(
			cfg,
			1,
			initSwitchFunc,
			SwitchFilterTimeout(5*time.Millisecond),
			SwitchPeerFilters(filters...),
		)
	)
	err := sw.Start()
	require.NoError(t, err)
	t.Cleanup(func() {
		if err := sw.Stop(); err != nil {
			t.Log(err)
		}
	})

	// simulate remote peer
	rp := newRemoteTCPPeer()
	rp.Start()
	defer rp.Stop()

	conn, err := sw.transport.Dial(*rp.Addr())
	if err != nil {
		t.Fatal(err)
	}

	p := wrapPeer(conn,
		rp.nodeInfo(),
		peerConfig{
			onPeerError:          sw.StopPeerForError,
			isPersistent:         sw.IsPeerPersistent,
			streamInfoByStreamID: sw.streamInfoByStreamID,
			metrics:              sw.metrics,
			outbound:             true,
		},
		rp.Addr())

	err = sw.addPeer(p)
	if _, ok := err.(tcp.ErrFilterTimeout); !ok {
		t.Errorf("expected ErrFilterTimeout")
	}
}

func TestSwitchPeerFilterDuplicate(t *testing.T) {
	sw := MakeSwitch(cfg, 1, initSwitchFunc)
	err := sw.Start()
	require.NoError(t, err)
	t.Cleanup(func() {
		if err := sw.Stop(); err != nil {
			t.Error(err)
		}
	})

	// simulate remote peer
	rp := newRemoteTCPPeer()
	rp.Start()
	defer rp.Stop()

	conn, err := sw.transport.Dial(*rp.Addr())
	if err != nil {
		t.Fatal(err)
	}

	p := wrapPeer(conn,
		rp.nodeInfo(),
		peerConfig{
			onPeerError:          sw.StopPeerForError,
			isPersistent:         sw.IsPeerPersistent,
			streamInfoByStreamID: sw.streamInfoByStreamID,
			metrics:              sw.metrics,
			outbound:             true,
		},
		rp.Addr())

	if err := sw.addPeer(p); err != nil {
		t.Fatal(err)
	}

	err = sw.addPeer(p)
	if errRej, ok := err.(ErrRejected); ok {
		if !errRej.IsDuplicate() {
			t.Errorf("expected peer to be duplicate. got %v", errRej)
		}
	} else {
		t.Errorf("expected ErrRejected, got %v", err)
	}
}

func assertNoPeersAfterTimeout(t *testing.T, sw *Switch, timeout time.Duration) {
	t.Helper()

	time.Sleep(timeout)
	if sw.Peers().Size() != 0 {
		t.Fatalf("Expected %v to not connect to some peers, got %d", sw, sw.Peers().Size())
	}
}

func TestSwitchStopsNonPersistentPeerOnError(t *testing.T) {
	sw := MakeSwitch(cfg, 1, initSwitchFunc)
	err := sw.Start()
	require.NoError(t, err)
	defer sw.Stop() //nolint:errcheck

	// simulate remote peer
	rp := newRemoteTCPPeer()
	rp.Start()
	defer rp.Stop()

	conn, err := sw.transport.Dial(*rp.Addr())
	require.NoError(t, err)

	p := wrapPeer(conn,
		rp.nodeInfo(),
		peerConfig{
			onPeerError:          sw.StopPeerForError,
			isPersistent:         sw.IsPeerPersistent,
			streamInfoByStreamID: sw.streamInfoByStreamID,
			metrics:              sw.metrics,
			outbound:             true,
		},
		rp.Addr())

	err = sw.addPeer(p)
	require.NoError(t, err)

	require.NotNil(t, sw.Peers().Get(rp.ID()))

	// simulate failure by closing connection
	err = conn.Close("simulate failure")
	require.NoError(t, err)

	assertNoPeersAfterTimeout(t, sw, 100*time.Millisecond)
	assert.False(t, p.IsRunning())
}

func TestSwitchStopPeerForError(t *testing.T) {
	s := httptest.NewServer(promhttp.Handler())
	defer s.Close()

	scrapeMetrics := func() string {
		resp, err := http.Get(s.URL)
		require.NoError(t, err)
		defer resp.Body.Close()
		buf, _ := io.ReadAll(resp.Body)
		return string(buf)
	}

	namespace, subsystem, name := config.TestInstrumentationConfig().Namespace, MetricsSubsystem, "peers"
	re := regexp.MustCompile(namespace + `_` + subsystem + `_` + name + ` ([0-9\.]+)`)
	peersMetricValue := func() float64 {
		matches := re.FindStringSubmatch(scrapeMetrics())
		f, _ := strconv.ParseFloat(matches[1], 64)
		return f
	}

	p2pMetrics := PrometheusMetrics(namespace)

	// make two connected switches
	sw1, sw2 := MakeSwitchPair(func(i int, sw *Switch) *Switch {
		// set metrics on sw1
		if i == 0 {
			opt := WithMetrics(p2pMetrics)
			opt(sw)
		}
		return initSwitchFunc(i, sw)
	})

	assert.Len(t, sw1.Peers().Copy(), 1)
	assert.EqualValues(t, 1, peersMetricValue())

	// send messages to the peer from sw1
	p := sw1.Peers().Copy()[0]
	err := p.Send(Envelope{
		ChannelID: 0x1,
		Message:   &p2pproto.Message{},
	})
	require.NoError(t, err)

	// stop sw2. this should cause the p to fail,
	// which results in calling StopPeerForError internally
	t.Cleanup(func() {
		if err := sw2.Stop(); err != nil {
			t.Error(err)
		}
	})

	// now call StopPeerForError explicitly, eg. from a reactor
	sw1.StopPeerForError(p, errors.New("some err"))

	require.Empty(t, len(sw1.Peers().Copy()), 0)
	assert.EqualValues(t, 0, peersMetricValue())
}

func TestSwitchReconnectsToOutboundPersistentPeer(t *testing.T) {
	sw := MakeSwitch(cfg, 1, initSwitchFunc)
	err := sw.Start()
	require.NoError(t, err)
	t.Cleanup(func() {
		if err := sw.Stop(); err != nil {
			t.Error(err)
		}
	})

	// 1. simulate failure by closing connection
	rp := newRemoteTCPPeer()
	rp.Start()
	defer rp.Stop()

	err = sw.AddPersistentPeers([]string{rp.Addr().String()})
	require.NoError(t, err)

	err = sw.DialPeerWithAddress(rp.Addr())
	require.NoError(t, err)
	require.NotNil(t, sw.Peers().Get(rp.ID()))

	p := sw.Peers().Copy()[0]
	err = p.(*peer).peerConn.Close("simulate failure")
	require.NoError(t, err)

	waitUntilSwitchHasAtLeastNPeers(sw, 1)
	assert.False(t, p.IsRunning())        // old peer instance
	assert.Equal(t, 1, sw.Peers().Size()) // new peer instance

	// 2. simulate first time dial failure
	rp = newRemoteTCPPeer()
	rp.Start()
	defer rp.Stop()

	conf := config.DefaultP2PConfig()
	conf.TestDialFail = true // will trigger a reconnect
	err = sw.addOutboundPeerWithConfig(rp.Addr(), conf)
	require.Error(t, err)
	// DialPeerWithAddres - sw.peerConfig resets the dialer
	waitUntilSwitchHasAtLeastNPeers(sw, 2)
	assert.Equal(t, 2, sw.Peers().Size())
}

func TestSwitchReconnectsToInboundPersistentPeer(t *testing.T) {
	sw := MakeSwitch(cfg, 1, initSwitchFunc)
	err := sw.Start()
	require.NoError(t, err)
	t.Cleanup(func() {
		if err := sw.Stop(); err != nil {
			t.Error(err)
		}
	})

	// 1. simulate failure by closing the connection
	rp := newRemoteTCPPeer()
	rp.Start()
	defer rp.Stop()

	err = sw.AddPersistentPeers([]string{rp.Addr().String()})
	require.NoError(t, err)

	conn, err := rp.Dial(sw.NetAddr())
	require.NoError(t, err)
	time.Sleep(50 * time.Millisecond)
	require.NotNil(t, sw.Peers().Get(rp.ID()))
	err = conn.Close("simulate failure")
	require.NoError(t, err)

	waitUntilSwitchHasAtLeastNPeers(sw, 1)
	assert.Equal(t, 1, sw.Peers().Size())
}

func TestSwitchDialPeersAsync(t *testing.T) {
	if testing.Short() {
		return
	}

	sw := MakeSwitch(cfg, 1, initSwitchFunc)
	err := sw.Start()
	require.NoError(t, err)
	t.Cleanup(func() {
		if err := sw.Stop(); err != nil {
			t.Error(err)
		}
	})

	rp := newRemoteTCPPeer()
	rp.Start()
	defer rp.Stop()

	err = sw.DialPeersAsync([]string{rp.Addr().String()})
	require.NoError(t, err)
	time.Sleep(dialRandomizerIntervalMilliseconds * time.Millisecond)
	require.NotNil(t, sw.Peers().Get(rp.ID()))
}

func waitUntilSwitchHasAtLeastNPeers(sw *Switch, n int) {
	for i := 0; i < 20; i++ {
		time.Sleep(250 * time.Millisecond)
		has := sw.Peers().Size()
		if has >= n {
			break
		}
	}
}

func TestSwitchFullConnectivity(t *testing.T) {
	switches := MakeConnectedSwitches(cfg, 3, initSwitchFunc, Connect2Switches)
	defer func() {
		for _, sw := range switches {
			t.Cleanup(func() {
				if err := sw.Stop(); err != nil {
					t.Error(err)
				}
			})
		}
	}()

	for i, sw := range switches {
		if sw.Peers().Size() != 2 {
			t.Fatalf("Expected each switch to be connected to 2 other, but %d switch only connected to %d", sw.Peers().Size(), i)
		}
	}
}

func TestSwitchAcceptRoutine(t *testing.T) {
	cfg.MaxNumInboundPeers = 5

	// Create some unconditional peers.
	const unconditionalPeersNum = 2
	var (
		unconditionalPeers   = make([]*remoteTCPPeer, unconditionalPeersNum)
		unconditionalPeerIDs = make([]string, unconditionalPeersNum)
	)
	for i := 0; i < unconditionalPeersNum; i++ {
		peer := newRemoteTCPPeer()
		peer.Start()
		unconditionalPeers[i] = peer
		unconditionalPeerIDs[i] = string(peer.ID())
	}

	// make switch
	sw := MakeSwitch(cfg, 1, initSwitchFunc)
	err := sw.AddUnconditionalPeerIDs(unconditionalPeerIDs)
	require.NoError(t, err)
	err = sw.Start()
	require.NoError(t, err)
	t.Cleanup(func() {
		err := sw.Stop()
		require.NoError(t, err)
	})

	// 0. check there are no peers
	assert.Equal(t, 0, sw.Peers().Size())

	// 1. check we connect up to MaxNumInboundPeers
	peers := make([]*remoteTCPPeer, 0)
	for i := 0; i < cfg.MaxNumInboundPeers; i++ {
		peer := newRemoteTCPPeer()
		peers = append(peers, peer)
		peer.Start()
		_, err := peer.Dial(sw.NetAddr())
		require.NoError(t, err)
		// spawn a reading routine to prevent connection from closing
		go func(s transport.Stream) {
			for {
				one := make([]byte, 1)
				_, err := s.Read(one)
				if err != nil {
					return
				}
			}
		}(peer.testStream)
	}
	time.Sleep(100 * time.Millisecond)
	assert.Equal(t, cfg.MaxNumInboundPeers, sw.Peers().Size())

	// 2. check we close new connections if we already have MaxNumInboundPeers peers
	peer := newRemoteTCPPeer()
	peer.Start()
	_, err = peer.Dial(sw.NetAddr())
	require.NoError(t, err)
	// check conn is closed
	one := make([]byte, 1)
	_ = peer.testStream.SetReadDeadline(time.Now().Add(10 * time.Millisecond))
	_, err = peer.testStream.Read(one)
	require.Error(t, err)
	assert.Equal(t, cfg.MaxNumInboundPeers, sw.Peers().Size())
	peer.Stop()

	// 3. check we connect to unconditional peers despite the limit.
	for _, peer := range unconditionalPeers {
		_, err := peer.Dial(sw.NetAddr())
		require.NoError(t, err)
		// spawn a reading routine to prevent connection from closing
		go func(s transport.Stream) {
			for {
				one := make([]byte, 1)
				_, err := s.Read(one)
				if err != nil {
					return
				}
			}
		}(peer.testStream)
	}
	time.Sleep(10 * time.Millisecond)
	assert.Equal(t, cfg.MaxNumInboundPeers+unconditionalPeersNum, sw.Peers().Size())

	for _, peer := range peers {
		peer.Stop()
	}
	for _, peer := range unconditionalPeers {
		peer.Stop()
	}
}

type errorTransport struct {
	acceptErr error
}

var _ transport.Transport = errorTransport{}

func (errorTransport) NetAddr() na.NetAddr {
	panic("not implemented")
}

func (et errorTransport) Accept() (transport.Conn, *na.NetAddr, error) {
	return nil, nil, et.acceptErr
}

func (errorTransport) Dial(na.NetAddr) (transport.Conn, error) {
	panic("not implemented")
}

func (errorTransport) Cleanup(transport.Conn) error {
	panic("not implemented")
}

func (errorTransport) UpdateStreamDescriptors([]transport.StreamDescriptor) {
	panic("not implemented")
}

func TestSwitchAcceptRoutineErrorCases(t *testing.T) {
	sw := NewSwitch(cfg, errorTransport{tcp.ErrFilterTimeout{}})
	assert.NotPanics(t, func() {
		err := sw.Start()
		require.NoError(t, err)
		err = sw.Stop()
		require.NoError(t, err)
	})

	// sw = NewSwitch(cfg, errorTransport{ErrRejected{conn: nil, err: errors.New("filtered"), isFiltered: true}})
	// assert.NotPanics(t, func() {
	// 	err := sw.Start()
	// 	require.NoError(t, err)
	// 	err = sw.Stop()
	// 	require.NoError(t, err)
	// })
	// TODO(melekes) check we remove our address from addrBook

	sw = NewSwitch(cfg, errorTransport{tcp.ErrTransportClosed{}})
	assert.NotPanics(t, func() {
		err := sw.Start()
		require.NoError(t, err)
		err = sw.Stop()
		require.NoError(t, err)
	})
}

// mockReactor checks that InitPeer never called before RemovePeer. If that's
// not true, InitCalledBeforeRemoveFinished will return true.
type mockReactor struct {
	*BaseReactor

	// atomic
	removePeerInProgress           uint32
	initCalledBeforeRemoveFinished uint32
}

func (r *mockReactor) RemovePeer(Peer, any) {
	atomic.StoreUint32(&r.removePeerInProgress, 1)
	defer atomic.StoreUint32(&r.removePeerInProgress, 0)
	time.Sleep(100 * time.Millisecond)
}

func (r *mockReactor) InitPeer(peer Peer) Peer {
	if atomic.LoadUint32(&r.removePeerInProgress) == 1 {
		atomic.StoreUint32(&r.initCalledBeforeRemoveFinished, 1)
	}

	return peer
}

func (r *mockReactor) InitCalledBeforeRemoveFinished() bool {
	return atomic.LoadUint32(&r.initCalledBeforeRemoveFinished) == 1
}

// see stopAndRemovePeer.
func TestSwitch_InitPeerIsNotCalledBeforeRemovePeer(t *testing.T) {
	// make reactor
	reactor := &mockReactor{}
	reactor.BaseReactor = NewBaseReactor("mockReactor", reactor)

	// make switch
	sw := MakeSwitch(cfg, 1, func(_ int, sw *Switch) *Switch {
		sw.AddReactor("mock", reactor)
		return sw
	})
	err := sw.Start()
	require.NoError(t, err)
	t.Cleanup(func() {
		if err := sw.Stop(); err != nil {
			t.Error(err)
		}
	})

	// add peer
	rp := newRemoteTCPPeer()
	rp.Start()
	defer rp.Stop()
	rpConn, err := rp.Dial(sw.NetAddr())
	require.NoError(t, err)

	// wait till the switch adds rp to the peer set, then stop the peer asynchronously
	for {
		time.Sleep(20 * time.Millisecond)
		if peer := sw.Peers().Get(rp.ID()); peer != nil {
			go sw.StopPeerForError(peer, "simulate failure")
			break
		}
	}

	// Close previous connection on the remote side too.
	//
	// This is necessary because rp.Dial may fail otherwise because rp may be not
	// quick enough in detecting that connection is dead.
	_ = rpConn.Close("simulate failure")
	// simulate peer reconnecting to us
	_, err = rp.Dial(sw.NetAddr())
	require.NoError(t, err)
	// wait till the switch adds rp to the peer set
	time.Sleep(50 * time.Millisecond)

	// make sure reactor.RemovePeer is finished before InitPeer is called
	assert.False(t, reactor.InitCalledBeforeRemoveFinished())
	assert.Equal(t, 1, sw.Peers().Size())
}

func makeSwitchForBenchmark(b *testing.B) *Switch {
	b.Helper()
	s1, s2 := MakeSwitchPair(initSwitchFunc)
	b.Cleanup(func() {
		if err := s2.Stop(); err != nil {
			b.Error(err)
		}
		if err := s1.Stop(); err != nil {
			b.Error(err)
		}
	})
	// Allow time for goroutines to boot up
	time.Sleep(1 * time.Second)
	return s1
}

func BenchmarkSwitchBroadcast(b *testing.B) {
	sw := makeSwitchForBenchmark(b)
	chMsg := &p2pproto.PexAddrs{
		Addrs: []p2pproto.NetAddress{
			{
				ID: "1",
			},
		},
	}

	b.ResetTimer()

	// Send random message from foo channel to another
	for i := 0; i < b.N; i++ {
		chID := byte(i % 4)
		sw.Broadcast(Envelope{ChannelID: chID, Message: chMsg})
	}
}

func BenchmarkSwitchTryBroadcast(b *testing.B) {
	sw := makeSwitchForBenchmark(b)
	chMsg := &p2pproto.PexAddrs{
		Addrs: []p2pproto.NetAddress{
			{
				ID: "1",
			},
		},
	}

	b.ResetTimer()

	// Send random message from foo channel to another
	for i := 0; i < b.N; i++ {
		chID := byte(i % 4)
		sw.TryBroadcast(Envelope{ChannelID: chID, Message: chMsg})
	}
}

func TestSwitchRemovalErr(t *testing.T) {
	sw1, sw2 := MakeSwitchPair(func(i int, sw *Switch) *Switch {
		return initSwitchFunc(i, sw)
	})
	require.Len(t, sw1.Peers().Copy(), 1)
	p := sw1.Peers().Copy()[0]

	sw2.StopPeerForError(p, errors.New("peer should error"))

	assert.Equal(t, sw2.peers.Add(p).Error(), ErrPeerRemoval{}.Error())
}

type remoteTCPPeer struct {
	privKey    crypto.PrivKey
	transport  *tcp.MultiplexTransport
	testStream transport.Stream
}

func newRemoteTCPPeer() *remoteTCPPeer {
	privKey := ed25519.GenPrivKey()
	nodeKey := nodekey.NodeKey{PrivKey: privKey}
	return &remoteTCPPeer{
		privKey:   privKey,
		transport: tcp.NewMultiplexTransport(nodeKey, tcpconn.DefaultMConnConfig()),
	}
}

func newRemoteTCPPeerWithPrivKey(privKey crypto.PrivKey) *remoteTCPPeer {
	nodeKey := nodekey.NodeKey{PrivKey: privKey}
	t := tcp.NewMultiplexTransport(nodeKey, tcpconn.DefaultMConnConfig())
	t.SetLogger(log.TestingLogger().With("peer", "remote"))
	return &remoteTCPPeer{
		privKey:   privKey,
		transport: t,
	}
}

func (rp *remoteTCPPeer) Addr() *na.NetAddr {
	na := rp.transport.NetAddr()
	return &na
}

func (rp *remoteTCPPeer) ID() nodekey.ID {
	return nodekey.PubKeyToID(rp.privKey.PubKey())
}

func (rp *remoteTCPPeer) Start() {
	id := nodekey.PubKeyToID(rp.privKey.PubKey())
	addr, err := na.NewFromString(string(id) + "@127.0.0.1:0")
	if err != nil {
		panic(err)
	}
	err = rp.transport.Listen(*addr)
	if err != nil {
		panic(err)
	}
	go rp.accept()
}

func (rp *remoteTCPPeer) Stop() {
	rp.transport.Close()
}

func (rp *remoteTCPPeer) Dial(addr *na.NetAddr) (transport.Conn, error) {
	c, err := rp.transport.Dial(*addr)
	if err != nil {
		return nil, err
	}

	_, err = handshake(rp.nodeInfo(), c.HandshakeStream(), time.Second)
	if err != nil {
		_ = c.Close(err.Error())
		return nil, err
	}

	rp.testStream, err = c.OpenStream(testCh, nil)
	if err != nil {
		_ = c.Close(err.Error())
		return nil, err
	}

	if mconn, ok := c.(*tcpconn.MConnection); ok {
		if err := mconn.Start(); err != nil {
			_ = c.Close(err.Error())
			return nil, fmt.Errorf("starting MConnection: %w", err)
		}
	}

	return c, err
}

func (rp *remoteTCPPeer) nodeInfo() ni.NodeInfo {
	la := rp.Addr()
	nodeInfo := testNodeInfo(rp.ID(), "remote_peer_"+la.String())
	nodeInfo.ListenAddr = la.DialString()
	nodeInfo.Channels = []byte{testCh}
	return nodeInfo
}

func (rp *remoteTCPPeer) accept() {
	for {
		c, _, err := rp.transport.Accept()
		if err != nil {
			return
		}

		_, err = handshake(rp.nodeInfo(), c.HandshakeStream(), time.Second)
		if err != nil {
			// Fixes TestSwitchFiltersOutItself.
			//
			// Without this timeout, rp will close its connection and the dialing
			// switch readLoop may fail before the handshake is done. Hence you will
			// see tcpconn.ErrNotRunning on Switch instead of ErrRejected. If we give
			// a little bit of time here, the dialing switch will always fail with
			// ErrRejected.
			time.Sleep(100 * time.Millisecond)
			_ = c.Close(err.Error())
			return
		}

		rp.testStream, err = c.OpenStream(testCh, nil)
		if err != nil {
			_ = c.Close(err.Error())
			return
		}

		if mconn, ok := c.(*tcpconn.MConnection); ok {
			if err := mconn.Start(); err != nil {
				_ = c.Close(err.Error())
				return
			}
		}
	}
}
