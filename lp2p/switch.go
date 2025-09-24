package lp2p

import (
	"fmt"
	"sync"

	"github.com/cometbft/cometbft/config"
	"github.com/cometbft/cometbft/libs/log"
	"github.com/cometbft/cometbft/libs/service"
	"github.com/cometbft/cometbft/p2p"
	"github.com/cosmos/gogoproto/proto"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/protocol"
	"github.com/pkg/errors"
)

// Switch represents p2p.Switcher alternative implementation based on go-libp2p.
// todo add comments to exported methods
// todo group unused methods
type Switch struct {
	service.BaseService

	config   *config.P2PConfig
	nodeKey  *p2p.NodeKey // our node private key
	nodeInfo p2p.NodeInfo // our node info

	host    *Host
	peerSet *PeerSet

	// reactorsByName represents [reactor_name => reactor] mapping
	reactorsByName map[string]p2p.Reactor

	// reactorsByProtocolID represents [protocol_id => reactor] mapping
	reactorsByProtocolID map[protocol.ID]p2p.Reactor

	// descriptorByProtocolID represents [protocol_id => channel_descriptor] mapping
	descriptorByProtocolID map[protocol.ID]*p2p.ChannelDescriptor

	// provisionedPeers represents set of peers that are added by reactors
	provisionedPeers map[p2p.ID]struct{}

	mu sync.RWMutex
}

// ReactorItem is a pair of name and reactor.
// Preserves order when adding.
type ReactorItem struct {
	Name    string
	Reactor p2p.Reactor
}

var _ p2p.Switcher = (*Switch)(nil)

var ErrUnsupportedPeerFormat = errors.New("unsupported peer format")

// NewSwitch constructs a new Switch.
func NewSwitch(
	cfg *config.P2PConfig,
	nodeKey *p2p.NodeKey,
	nodeInfo p2p.NodeInfo,
	host *Host,
	reactors []ReactorItem,
	logger log.Logger,
) *Switch {
	s := &Switch{
		config:   cfg,
		nodeInfo: nodeInfo,
		nodeKey:  nodeKey,

		host:    host,
		peerSet: NewPeerSet(host, logger),

		reactorsByName:         make(map[string]p2p.Reactor),
		reactorsByProtocolID:   make(map[protocol.ID]p2p.Reactor),
		descriptorByProtocolID: make(map[protocol.ID]*p2p.ChannelDescriptor),
	}

	base := service.NewBaseService(logger, "LibP2P Switch", s)
	s.BaseService = *base

	for _, el := range reactors {
		s.AddReactor(el.Name, el.Reactor)
	}

	return s
}

//--------------------------------
// BaseService methods
//--------------------------------

func (s *Switch) OnStart() error {
	s.Logger.Info("Starting LibP2PSwitch")

	for name, reactor := range s.reactorsByName {
		err := reactor.OnStart()
		if err != nil {
			return fmt.Errorf("failed to start reactor %s: %w", name, err)
		}
	}

	return nil
}

func (s *Switch) OnStop() {
	s.Logger.Info("Stopping LibP2PSwitch")

	for name, reactor := range s.reactorsByName {
		if err := reactor.Stop(); err != nil {
			s.Logger.Error("failed to stop reactor", "name", name, "err", err)
		}
	}

	if err := s.host.Network().Close(); err != nil {
		s.Logger.Error("failed to close network", "err", err)
	}

	if err := s.host.Peerstore().Close(); err != nil {
		s.Logger.Error("failed to close peerstore", "err", err)
	}
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
	reactor, exists := s.reactorsByName[name]

	return reactor, exists
}

// AddReactor adds the given reactor to the switch.
// NOTE: Not goroutine safe.
func (s *Switch) AddReactor(name string, reactor p2p.Reactor) p2p.Reactor {
	for i := range reactor.GetChannels() {
		var (
			channelDescriptor = reactor.GetChannels()[i]
			protocolID        = ProtocolID(channelDescriptor.ID)
		)

		// Comet compatibility: ensure channelID is unique across all reactors
		if _, ok := s.reactorsByProtocolID[protocolID]; ok {
			err := fmt.Errorf("adding reactor %q: protocol %q is already registered", name, protocolID)
			panic(err)
		}

		s.host.SetStreamHandler(protocolID, s.handleStream)
	}

	s.reactorsByName[name] = reactor
	reactor.SetSwitch(s)

	return reactor
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

// AddPersistentPeers addrs peers in a format of id@ip:port
func (s *Switch) AddPersistentPeers(addrs []string) error {
	// since lib-p2p relies on multiaddr format, we can't use it
	return ErrUnsupportedPeerFormat
}

// AddPrivatePeerIDs ids peers in a format of Comet peer id
func (s *Switch) AddPrivatePeerIDs(ids []string) error {
	// since lib-p2p relies on multiaddr format, we can't use it
	return ErrUnsupportedPeerFormat
}

// AddUnconditionalPeerIDs ids peers in a format of Comet peer id
func (s *Switch) AddUnconditionalPeerIDs(ids []string) error {
	// since lib-p2p relies on multiaddr format, we can't use it
	return ErrUnsupportedPeerFormat
}

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
	s.Logger.Info("Stopping peer for error", "peer", peer, "reason", reason)

	p, ok := peer.(*Peer)
	if !ok {
		s.Logger.Error("Peer is not a lp2p.Peer", "peer", peer, "reason", reason)
		return
	}

	if err := s.deprovisionPeer(p, reason); err != nil {
		s.Logger.Error("Failed to deprovision peer", "peer", peer, "err", err)
	}
}

func (s *Switch) IsDialingOrExistingAddress(addr *p2p.NetAddress) bool {
	s.logUnimplemented("IsDialingOrExistingAddress")
	return false
}

func (s *Switch) IsPeerPersistent(_ *p2p.NetAddress) bool {
	s.logUnimplemented("IsPeerPersistent")
	return false
}

func (s *Switch) IsPeerUnconditional(id p2p.ID) bool {
	// todo: add support for unconditional peers (used by mempool reactor)
	return false
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

	s.peerSet.ForEach(func(p p2p.Peer) {
		go p.Send(e)
	})
}

func (s *Switch) TryBroadcast(e p2p.Envelope) {
	s.Logger.Debug("TryBroadcast", "channel", e.ChannelID)

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

	defer func() {
		if r := recover(); r != nil {
			s.Logger.Error("Panic in Switch.handleStream",
				"peer", peerID,
				"protocol", protocolID,
				"panic", r,
			)
			_ = stream.Reset()
		}
	}()

	// 1. Retrieve the reactor with channel descriptor
	reactor, ok := s.reactorsByProtocolID[protocolID]
	if !ok {
		// should not happen
		s.Logger.Error("Unknown protocol", "protocol", protocolID)
		_ = stream.Reset()
		return
	}

	descriptor, ok := s.descriptorByProtocolID[protocolID]
	if !ok {
		// should not happen
		s.Logger.Error("Unknown protocol descriptor", "protocol", protocolID)
		_ = stream.Reset()
		return
	}

	// 2. Retrieve the peer from the peerSet
	peer := s.peerSet.Get(p2p.ID(peerID))
	if peer == nil {
		s.Logger.Error("Unable to get peer from peerSet", "peer", peerID)
		return
	}

	// 3. Read the stream so we can "release" it on another end
	payload, err := StreamReadClose(stream)
	if err != nil {
		s.Logger.Error("Failed to read payload", "protocol", protocolID, "err", err)
		return
	}

	msg, err := unmarshalProto(descriptor, payload)
	if err != nil {
		s.Logger.Error("Failed to unmarshal message", "protocol", protocolID, "err", err)
		return
	}

	// 4. Ensure peer is provisioned
	if err := s.ensurePeerProvisioned(peer); err != nil {
		s.Logger.Error("Failed to ensure peer is provisioned", "peer", peerID, "err", err)
		return
	}

	s.Logger.Debug("Received stream envelope", "peer", peerID, "protocol", protocolID, "message", msg)

	// todo metrics

	reactor.Receive(p2p.Envelope{
		Src:       peer,
		ChannelID: descriptor.ID,
		Message:   msg,
	})
}

func (s *Switch) ensurePeerProvisioned(peer p2p.Peer) error {
	s.mu.RLock()
	_, peerProvisioned := s.provisionedPeers[peer.ID()]
	s.mu.RUnlock()

	// noop
	if peerProvisioned {
		return nil
	}

	// should not happen
	p, ok := peer.(*Peer)
	if !ok {
		return errors.New("peer is not a lp2p.Peer")
	}

	return s.provisionPeer(p)
}

func (s *Switch) provisionPeer(peer *Peer) error {
	// todo filter peers ? we should use ConnGater instead

	logger := s.Logger.With("peer", peer.addrInfo.String())

	peer.SetLogger(logger)

	// should not happen
	if !s.IsRunning() {
		return errors.New("switch is not running")
	}

	// note: order is not guaranteed (however we don't care in this case)
	for _, reactor := range s.reactorsByName {
		reactor.InitPeer(peer)
		reactor.AddPeer(peer)
	}

	// note that we don't need to start Peer as with Comet's Peer
	// because it's a thin wrapper and it doesn't handle streams

	s.mu.Lock()
	s.provisionedPeers[peer.ID()] = struct{}{}
	s.mu.Unlock()

	return nil
}

func (s *Switch) deprovisionPeer(peer *Peer, reason any) error {
	key := peer.ID()
	id := peer.addrInfo.ID

	s.peerSet.Remove(key)

	for _, reactor := range s.reactorsByName {
		reactor.RemovePeer(peer, reason)
	}

	if err := s.host.Network().ClosePeer(id); err != nil {
		s.Logger.Error("Failed to close peer", "peer", peer, "err", err)
	}

	s.peerSet.Remove(key)

	s.mu.Lock()
	delete(s.provisionedPeers, key)
	s.mu.Unlock()

	return nil
}

func marshalProto(msg proto.Message) ([]byte, error) {
	// comet compatibility
	// @see p2p/peer.go (*peer).send()
	if w, ok := msg.(p2p.Wrapper); ok {
		msg = w.Wrap()
	}

	payload, err := proto.Marshal(msg)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to marshal proto")
	}

	return payload, nil
}

func unmarshalProto(descriptor *p2p.ChannelDescriptor, payload []byte) (proto.Message, error) {
	var (
		msg = proto.Clone(descriptor.MessageType)
		err = proto.Unmarshal(payload, msg)
	)

	if err != nil {
		return nil, errors.Wrapf(err, "failed to unmarshal message")
	}

	// comet compatibility
	// @see p2p/peer.go createMConnection()
	if w, ok := msg.(p2p.Unwrapper); ok {
		msg, err = w.Unwrap()
		if err != nil {
			return nil, errors.Wrapf(err, "failed to unwrap message")
		}
	}

	return msg, nil
}
