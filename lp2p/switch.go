package lp2p

import (
	"context"
	"fmt"
	"math/rand"
	"runtime/debug"
	"sync"
	"sync/atomic"
	"time"

	"github.com/cometbft/cometbft/libs/log"
	"github.com/cometbft/cometbft/libs/service"
	"github.com/cometbft/cometbft/p2p"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/peerstore"
	"github.com/libp2p/go-libp2p/core/protocol"
	"github.com/pkg/errors"
)

// Switch represents p2p.Switcher alternative implementation based on go-libp2p.
// todo add comments to exported methods
// todo group unused methods
type Switch struct {
	service.BaseService

	nodeInfo p2p.NodeInfo // our node info

	host    *Host
	peerSet *PeerSet

	reactors *reactorSet

	metrics *p2p.Metrics

	// active is used to track if the switch has started
	// BaseService has similar field, but it triggered BEFORE OnStart().
	// This leads to concurrent peers provisioning between bootstrapping peers and accepting incoming messages
	active atomic.Bool
}

// SwitchReactor is a pair of name and reactor.
// Preserves order when adding.
type SwitchReactor struct {
	p2p.Reactor
	Name string
}

const MaxReconnectBackoff = 5 * time.Minute

var _ p2p.Switcher = (*Switch)(nil)

var ErrUnsupportedPeerFormat = errors.New("unsupported peer format")

// NewSwitch constructs a new Switch.
func NewSwitch(
	nodeInfo p2p.NodeInfo,
	host *Host,
	reactors []SwitchReactor,
	metrics *p2p.Metrics,
	logger log.Logger,
) (*Switch, error) {
	s := &Switch{
		nodeInfo: nodeInfo,

		host:    host,
		peerSet: NewPeerSet(host, metrics, logger),

		metrics: metrics,

		active: atomic.Bool{},
	}

	base := service.NewBaseService(logger, "LibP2P Switch", s)
	s.BaseService = *base

	s.reactors = newReactorSet(s)

	for _, item := range reactors {
		if err := s.reactors.Add(item.Reactor, item.Name); err != nil {
			return nil, errors.Wrapf(err, "failed to add %q reactor", item.Name)
		}
	}

	return s, nil
}

//--------------------------------
// BaseService methods
//--------------------------------

func (s *Switch) OnStart() error {
	s.Logger.Info("Starting lib-p2p switch")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	protocolHandler := func(protocolID protocol.ID) {
		s.host.SetStreamHandler(protocolID, s.handleStream)
	}

	// 1. start reactors
	err := s.reactors.Start(protocolHandler)
	if err != nil {
		return fmt.Errorf("failed to start reactors: %w", err)
	}

	// 2. register peer failure handler
	s.host.AddPeerFailureHandler(func(id peer.ID, err error) {
		key := peerIDToKey(id)
		peer := s.peerSet.Get(key)
		s.StopPeerForError(peer, err)
	})

	// 3. connect bootstrap peers
	bootstrapPeers := s.host.BootstrapPeers()

	s.Logger.Info("Connecting to bootstrap peers", "count", len(bootstrapPeers))

	for _, bp := range bootstrapPeers {
		opts := PeerAddOptions{
			Private:       bp.Private,
			Persistent:    bp.Persistent,
			Unconditional: bp.Unconditional,
			OnBeforeStart: s.reactors.InitPeer,
			OnAfterStart:  s.reactors.AddPeer,
			OnStartFailed: s.reactors.RemovePeer,
		}

		err := s.bootstrapPeer(ctx, bp.AddrInfo, opts)
		if err != nil {
			s.Logger.Error("Unable to add bootstrap peer", "peer_id", bp.AddrInfo.String(), "err", err)
			go s.reconnectPeer(bp.AddrInfo, MaxReconnectBackoff, opts)
			continue
		}
	}

	s.active.Store(true)

	return nil
}

func (s *Switch) OnStop() {
	s.Logger.Info("Stopping LibP2PSwitch")

	s.reactors.Stop()
	s.peerSet.RemoveAll(PeerRemovalOptions{Reason: "switch stopped"})

	if err := s.host.Network().Close(); err != nil {
		s.Logger.Error("failed to close network", "err", err)
	}

	if err := s.host.Peerstore().Close(); err != nil {
		s.Logger.Error("failed to close peerstore", "err", err)
	}

	s.active.Store(false)
}

func (s *Switch) NodeInfo() p2p.NodeInfo {
	return s.nodeInfo
}

func (s *Switch) Log() log.Logger {
	return s.Logger
}

//--------------------------------
// ReactorManager methods
//--------------------------------

func (s *Switch) Reactor(name string) (p2p.Reactor, bool) {
	return s.reactors.GetByName(name)
}

// AddReactor adds the given reactor to the switch.
// NOTE: Not goroutine safe.
func (s *Switch) AddReactor(name string, reactor p2p.Reactor) p2p.Reactor {
	// used only by CustomReactors
	s.logUnimplemented("AddReactor")

	return nil
}

func (s *Switch) RemoveReactor(_ string, _ p2p.Reactor) {
	// used only by CustomReactors
	s.logUnimplemented("RemoveReactor")
}

// --------------------------------
// PeerManager methods
// --------------------------------

func (s *Switch) Peers() p2p.IPeerSet {
	return s.peerSet
}

func (s *Switch) NumPeers() (outbound, inbound, dialing int) {
	for _, c := range s.host.Network().Conns() {
		switch c.Stat().Direction {
		case network.DirInbound:
			inbound++
		case network.DirOutbound:
			outbound++
		}
	}

	// todo note we don't count dialing peers here

	return outbound, inbound, dialing
}

func (s *Switch) MaxNumOutboundPeers() int {
	// used only by PEX
	s.logUnimplemented("MaxNumOutboundPeers")

	return 0
}

func (s *Switch) AddPersistentPeers(addrs []string) error    { return ErrUnsupportedPeerFormat }
func (s *Switch) AddPrivatePeerIDs(ids []string) error       { return ErrUnsupportedPeerFormat }
func (s *Switch) AddUnconditionalPeerIDs(ids []string) error { return ErrUnsupportedPeerFormat }

func (s *Switch) DialPeerWithAddress(_ *p2p.NetAddress) error {
	// used only by PEX
	s.logUnimplemented("DialPeerWithAddress")

	return nil
}

func (s *Switch) DialPeersAsync(peers []string) error {
	s.logUnimplemented("DialPeersAsync", "peers", peers)

	return nil
}

func (s *Switch) StopPeerGracefully(_ p2p.Peer) {
	// used only by PEX
	s.logUnimplemented("StopPeerGracefully")
}

func (s *Switch) StopPeerForError(peer p2p.Peer, reason any) {
	// should not happen
	p, ok := peer.(*Peer)
	if !ok {
		return
	}

	pid := p.ID()

	removalOpts := PeerRemovalOptions{
		Reason:      reason,
		OnAfterStop: s.reactors.RemovePeer,
	}

	if err := s.peerSet.Remove(pid, removalOpts); err != nil {
		s.Logger.Error("Failed to remove peer", "peer_id", pid, "err", err)
		return
	}

	// todo, actually, for persistent peers we can skip this step,
	// but explicitly closing might cleanup some conns/resources
	if err := s.host.Network().ClosePeer(p.addrInfo.ID); err != nil {
		// tolerate this error.
		s.Logger.Error("Failed to close peer", "peer_id", pid, "err", err)
	}

	// reconnect logic
	shouldReconnect := false

	if p.IsPersistent() {
		shouldReconnect = true
		s.Logger.Debug("Will reconnect to peer", "peer_id", pid, "err", reason)
	} else if errTransient, ok := TransientErrorFromAny(reason); ok {
		shouldReconnect = true
		s.Logger.Debug("Will reconnect to peer after transient error", "peer_id", pid, "err", errTransient.Err)
	}

	if !shouldReconnect {
		return
	}

	go s.reconnectPeer(p.AddrInfo(), MaxReconnectBackoff, PeerAddOptions{
		Persistent:    p.IsPersistent(),
		Unconditional: p.IsUnconditional(),
		Private:       p.IsPrivate(),
		OnBeforeStart: s.reactors.InitPeer,
		OnAfterStart:  s.reactors.AddPeer,
		OnStartFailed: s.reactors.RemovePeer,
	})
}

func (s *Switch) IsDialingOrExistingAddress(addr *p2p.NetAddress) bool {
	s.logUnimplemented("IsDialingOrExistingAddress")
	return false
}

func (s *Switch) IsPeerPersistent(netAddr *p2p.NetAddress) bool {
	p := s.peerSet.Get(netAddr.ID)
	if p == nil {
		return false
	}

	return p.(*Peer).IsPersistent()
}

func (s *Switch) IsPeerUnconditional(id p2p.ID) bool {
	p := s.peerSet.Get(id)
	if p == nil {
		return false
	}

	return p.(*Peer).IsUnconditional()
}

func (s *Switch) MarkPeerAsGood(_ p2p.Peer) {
	// used by consensus reactor
	s.logUnimplemented("MarkPeerAsGood")
}

//--------------------------------
// Broadcaster methods
//--------------------------------

func (s *Switch) Broadcast(e p2p.Envelope) chan bool {
	s.Logger.Debug("Broadcast", "channel", e.ChannelID)

	e.Message = newPreMarshaledMessage(e.Message)

	var wg sync.WaitGroup
	successChan := make(chan bool, s.peerSet.Size())

	s.peerSet.ForEach(func(p p2p.Peer) {
		wg.Add(1)

		go func(p p2p.Peer) {
			defer wg.Done()

			success := p.Send(e)
			select {
			case successChan <- success:
			default:
				// Skip. This means peer set changed
				// between Size() and ForEach() calls.
			}
		}(p)
	})

	go func() {
		wg.Wait()
		close(successChan)
	}()

	return successChan
}

func (s *Switch) BroadcastAsync(e p2p.Envelope) {
	s.Logger.Debug("BroadcastAsync", "channel", e.ChannelID)

	e.Message = newPreMarshaledMessage(e.Message)

	s.peerSet.ForEach(func(p p2p.Peer) {
		go p.Send(e)
	})
}

func (s *Switch) TryBroadcast(e p2p.Envelope) {
	s.Logger.Debug("TryBroadcast", "channel", e.ChannelID)

	e.Message = newPreMarshaledMessage(e.Message)

	s.peerSet.ForEach(func(p p2p.Peer) {
		go p.TrySend(e)
	})
}

func (s *Switch) logUnimplemented(method string, kv ...any) {
	s.Logger.Info(
		"Unimplemented LibP2PSwitch method called",
		append(kv, "method", method)...,
	)
}

func (s *Switch) handleStream(stream network.Stream) {
	var (
		peerID     = stream.Conn().RemotePeer()
		protocolID = stream.Protocol()
	)

	if !s.isActive() {
		s.Log().Debug(
			"Ignoring stream from inactive switch",
			"peer_id", peerID.String(),
			"protocol", protocolID,
		)
		_ = stream.Reset()
		return
	}

	defer func() {
		if r := recover(); r != nil {
			s.Logger.Error(
				"Panic in (*lp2p.Switch).handleStream",
				"peer_id", peerID.String(),
				"protocol", protocolID,
				"panic", r,
				"stack", string(debug.Stack()),
			)
			_ = stream.Reset()
		}
	}()

	// 1. Retrieve the reactor with channel descriptor
	proto, reactor, err := s.reactors.getReactorWithProtocol(protocolID)
	if err != nil {
		// should not happen
		s.Logger.Error("Unknown protocol descriptor", "protocol", protocolID)
		_ = stream.Reset()
		return
	}

	// 2. Read the stream so we can "release" it on another end
	payload, err := StreamReadClose(stream)
	if err != nil {
		s.Logger.Error("Failed to read payload", "protocol", protocolID, "err", err)
		return
	}

	msg, err := unmarshalProto(proto.descriptor, payload)
	if err != nil {
		s.Logger.Error("Failed to unmarshal message", "protocol", protocolID, "err", err)
		return
	}

	// 3. Retrieve the peer from the peerSet (or provision if it's not)
	peer, err := s.resolvePeer(peerID)
	if err != nil {
		s.Logger.Error("Failed to resolve peer", "peer_id", peerID.String(), "err", err)
		return
	}

	var (
		// peer id is for receive metrics
		peerStr     = peerID.String()
		messageType = protoTypeName(msg)
		payloadLen  = float64(len(payload))
		labels      = []string{
			"peer_id", peerStr,
			"chID", fmt.Sprintf("%#x", proto.descriptor.ID),
		}
	)

	s.metrics.PeerReceiveBytesTotal.With(labels...).Add(payloadLen)
	s.metrics.MessageReceiveBytesTotal.With("message_type", messageType).Add(payloadLen)

	s.Logger.Debug(
		"Received stream envelope. Submitting to reactor",
		"peer_id", peerID.String(),
		"protocol", protocolID,
		"message_type", messageType,
		"payload_len", payloadLen,
	)

	envelope := p2p.Envelope{
		Src:       peer,
		ChannelID: proto.descriptor.ID,
		Message:   msg,
	}

	priority := proto.descriptor.Priority

	s.reactors.Receive(reactor.name, messageType, envelope, priority)
}

func (s *Switch) resolvePeer(id peer.ID) (p2p.Peer, error) {
	key := peerIDToKey(id)

	// peer exists (99% of the time)
	if peer := s.peerSet.Get(key); peer != nil {
		return peer, nil
	}

	addrInfo := s.host.Peerstore().PeerInfo(id)
	if len(addrInfo.Addrs) == 0 {
		return nil, errors.New("peer has no addresses in peerstore")
	}

	// let's try to provision it
	opts := PeerAddOptions{
		Private:       false,
		Persistent:    false,
		Unconditional: false,
		OnBeforeStart: s.reactors.InitPeer,
		OnAfterStart:  s.reactors.AddPeer,
		OnStartFailed: s.reactors.RemovePeer,
	}

	peer, err := s.peerSet.Add(addrInfo, opts)
	switch {
	case errors.Is(err, ErrPeerExists):
		// tolerate two concurrent peer additions
		if p := s.peerSet.Get(key); p != nil {
			return p, nil
		}

		// should not happen
		return nil, errors.Wrap(err, "peer exists but not found")
	case err != nil:
		return nil, errors.Wrap(err, "unable to add peer")
	default:
		return peer, nil
	}
}

// bootstrapPeer connects a bootstrap peer to the host. should be used ONLY during switch start.
func (s *Switch) bootstrapPeer(ctx context.Context, addrInfo peer.AddrInfo, opts PeerAddOptions) error {
	if addrInfo.ID == s.host.ID() {
		s.Logger.Info("Ignoring connection to self")
		return nil
	}

	pid := addrInfo.ID.String()

	s.Logger.Info(
		"Connecting to peer",
		"peer_id", pid,
		"addr_info", addrInfo.String(),
		"persistent", opts.Persistent,
		"unconditional", opts.Unconditional,
		"private", opts.Private,
	)

	if err := s.host.Connect(ctx, addrInfo); err != nil {
		return errors.Wrap(err, "unable to connect to peer")
	}

	if _, err := s.peerSet.Add(addrInfo, opts); err != nil {
		return errors.Wrap(err, "unable to add peer")
	}

	for _, addr := range addrInfo.Addrs {
		// if this a DNS address, save it to the peerstore, so in case of IP address change,
		// we can use the DNS address to reconnect during message sending.
		if IsDNSAddr(addr) {
			s.host.Peerstore().AddAddr(addrInfo.ID, addr, peerstore.PermanentAddrTTL)
		}
	}

	// after connecting, there might be more than
	// one address (eg `/dns/...` + resolved `/ipv4/...`)
	addresses := s.host.multiAddrStrByID(addrInfo.ID)

	s.Logger.Info("Connected to peer", "peer_id", pid, "addresses", addresses)

	go s.pingPeer(addrInfo)

	return nil
}

// reconnectPeer reconnects persistent peers back to the host.
// uses exponential backoff to reconnect.
func (s *Switch) reconnectPeer(addrInfo peer.AddrInfo, backoffMax time.Duration, opts PeerAddOptions) {
	defer func() {
		if r := recover(); r != nil {
			s.Logger.Error("Panic in (*lp2p.Switch).reconnectTo", "panic", r)
		}
	}()

	backoff := 1 * time.Second
	sleep := func() {
		jitter := time.Duration(rand.Intn(100)) * time.Millisecond
		time.Sleep(backoff + jitter)

		backoff *= 2
		if backoffMax > 0 && backoff > backoffMax {
			backoff = backoffMax
		}
	}

	var (
		ctx   = network.WithDialPeerTimeout(context.Background(), 3*time.Second)
		pid   = addrInfo.ID.String()
		start = time.Now()
	)

	for {
		if !s.isActive() {
			return
		}

		s.Logger.Info(
			"Reconnecting to peer",
			"peer_id", pid,
			"private", opts.Private,
			"persistent", opts.Persistent,
			"unconditional", opts.Unconditional,
		)

		// 1. ensure connection (dial or noop if already connected)
		if err := s.host.Connect(ctx, addrInfo); err != nil {
			s.Logger.Error(
				"Failed to reconnect to peer",
				"peer_id", pid,
				"err", err,
				"backoff", backoff.String(),
			)

			sleep()
			continue
		}

		// 2. add peer to the peer set
		_, err := s.peerSet.Add(addrInfo, opts)
		if err != nil && !errors.Is(err, ErrPeerExists) {
			s.Logger.Error(
				"Failed to add peer after reconnection",
				"peer_id", pid,
				"err", err,
				"backoff", backoff.String(),
			)
			sleep()
			continue
		}

		var (
			elapsed   = time.Since(start)
			addresses = s.host.multiAddrStrByID(addrInfo.ID)
		)

		s.Logger.Info(
			"Reconnected to peer",
			"peer_id", pid,
			"addresses", addresses,
			"elapsed", elapsed.String(),
		)

		go s.pingPeer(addrInfo)

		return
	}
}

// pingPeer pings peers and logs RTT latency (blocking)
// Keep in might that ping service might be disabled on the counterparty side.
func (s *Switch) pingPeer(addrInfo peer.AddrInfo) {
	const timeout = 5 * time.Second

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	var (
		pid       = addrInfo.ID.String()
		addresses = s.host.multiAddrStrByID(addrInfo.ID)
	)

	rtt, err := s.host.Ping(ctx, addrInfo)
	if err != nil {
		s.Logger.Error("Failed to ping peer", "peer_id", pid, "addresses", addresses, "err", err)
		return
	}

	s.Logger.Info("Ping", "peer_id", pid, "addresses", addresses, "rtt", rtt.String())
}

func (s *Switch) isActive() bool {
	return s.active.Load()
}
