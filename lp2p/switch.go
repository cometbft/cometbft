package lp2p

import (
	"context"
	"fmt"
	"runtime/debug"
	"sync"
	"time"

	"github.com/cometbft/cometbft/config"
	"github.com/cometbft/cometbft/libs/log"
	"github.com/cometbft/cometbft/libs/service"
	"github.com/cometbft/cometbft/p2p"
	"github.com/libp2p/go-libp2p/core/event"
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

	reactors []ReactorItem

	// reactorsByName represents [reactor_name => reactor] mapping
	reactorsByName map[string]p2p.Reactor

	// reactorsByProtocolID represents [protocol_id => reactor] mapping
	reactorsByProtocolID map[protocol.ID]p2p.Reactor

	// descriptorByProtocolID represents [protocol_id => channel_descriptor] mapping
	descriptorByProtocolID map[protocol.ID]*p2p.ChannelDescriptor

	// provisionedPeers represents set of peers that are added by reactors
	// todo should it live within peerSet?
	provisionedPeers map[p2p.ID]struct{}

	eventBusSubscription event.Subscription

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
) (*Switch, error) {
	s := &Switch{
		config:   cfg,
		nodeInfo: nodeInfo,
		nodeKey:  nodeKey,

		host:    host,
		peerSet: NewPeerSet(host, logger),

		reactors:               make([]ReactorItem, 0, len(reactors)),
		reactorsByName:         make(map[string]p2p.Reactor),
		reactorsByProtocolID:   make(map[protocol.ID]p2p.Reactor),
		descriptorByProtocolID: make(map[protocol.ID]*p2p.ChannelDescriptor),
		provisionedPeers:       make(map[p2p.ID]struct{}),
	}

	base := service.NewBaseService(logger, "LibP2P Switch", s)
	s.BaseService = *base

	for _, el := range reactors {
		s.AddReactor(el.Name, el.Reactor)
	}

	eventTypes := []any{
		&event.EvtPeerConnectednessChanged{},
	}

	sub, err := s.host.EventBus().Subscribe(eventTypes)
	if err != nil {
		return nil, errors.Wrap(err, "unable to subscribe to event bus")
	}

	s.eventBusSubscription = sub

	return s, nil
}

//--------------------------------
// BaseService methods
//--------------------------------

func (s *Switch) OnStart() error {
	s.Logger.Info("Starting LibP2PSwitch")

	go s.eventListener()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	ConnectPeers(ctx, s.host, s.host.ConfigPeers())

	for _, el := range s.reactors {
		name, reactor := el.Name, el.Reactor

		s.Logger.Info("Starting reactor", "reactor", name)

		if err := reactor.Start(); err != nil {
			return fmt.Errorf("failed to start reactor %s: %w", name, err)
		}
	}

	peers := s.peerSet.Copy()
	for _, peer := range peers {
		if err := s.ensurePeerProvisioned(peer); err != nil {
			s.Logger.Error("Failed to provision peer", "peer_id", peer.ID(), "err", err)
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

	if err := s.eventBusSubscription.Close(); err != nil {
		s.Logger.Error("failed to close event bus subscription", "err", err)
	}

	// todo disconnect from config peers!
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
	s.Logger.Info("Adding reactor", "name", name)

	// set reactor's channels
	for i := range reactor.GetChannels() {
		var (
			channelDescriptor = reactor.GetChannels()[i]
			protocolID        = ProtocolID(channelDescriptor.ID)
		)

		// Ensure channelID is unique across all reactors
		if _, ok := s.reactorsByProtocolID[protocolID]; ok {
			err := fmt.Errorf("adding reactor %q: protocol %q is already registered", name, protocolID)
			panic(err)
		}

		s.reactorsByProtocolID[protocolID] = reactor
		s.descriptorByProtocolID[protocolID] = channelDescriptor

		s.host.SetStreamHandler(protocolID, s.handleStream)
	}

	// set reactor itself
	s.reactors = append(s.reactors, ReactorItem{Name: name, Reactor: reactor})
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
			s.Logger.Error(
				"Panic in (*lp2p.Switch).handleStream",
				"peer", peerID,
				"protocol", protocolID,
				"panic", r,
				"stack", string(debug.Stack()),
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
	peer := s.peerSet.Get(peerIDToKey(peerID))
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

	s.Logger.Debug(
		"Received stream envelope",
		"peer", peerID,
		"protocol", protocolID,
		"message_type", log.NewLazySprintf("%T", msg),
		"message", msg,
	)

	// todo metrics

	reactor.Receive(p2p.Envelope{
		Src:       peer,
		ChannelID: descriptor.ID,
		Message:   msg,
	})
}

func (s *Switch) ensurePeerProvisioned(peer p2p.Peer) error {
	s.mu.RLock()
	_, exists := s.provisionedPeers[peer.ID()]
	s.mu.RUnlock()

	// noop
	if exists {
		return nil
	}

	// should not happen
	p, ok := peer.(*Peer)
	if !ok {
		return errors.New("peer is not a lp2p.Peer")
	}

	return s.provisionPeer(p)
}

// effectively called once per peer. note that we don't need to start Peer as with Comet's Peer
// because it's a thin wrapper and it doesn't handle streams
func (s *Switch) provisionPeer(peer *Peer) error {
	// todo filter peers ? we should use ConnGater instead

	// should not happen
	if !s.IsRunning() {
		return errors.New("switch is not running")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// noop - might be the case during init phase.
	if _, ok := s.provisionedPeers[peer.ID()]; ok {
		return nil
	}

	s.Logger.Info("Provisioning peer", "peer_id", peer.ID())

	for _, el := range s.reactors {
		el.Reactor.InitPeer(peer)
	}

	if err := peer.Start(); err != nil {
		return errors.Wrap(err, "failed to start peer")
	}

	for _, el := range s.reactors {
		el.Reactor.AddPeer(peer)
	}

	s.provisionedPeers[peer.ID()] = struct{}{}

	return nil
}

func (s *Switch) deprovisionPeer(peer *Peer, reason any) error {
	key := peer.ID()
	id := peer.addrInfo.ID

	s.mu.Lock()
	defer s.mu.Unlock()

	s.peerSet.Remove(key)

	if err := peer.Stop(); err != nil {
		return errors.Wrap(err, "failed to stop peer")
	}

	for _, reactor := range s.reactorsByName {
		reactor.RemovePeer(peer, reason)
	}

	if err := s.host.Network().ClosePeer(id); err != nil {
		s.Logger.Error("Failed to close peer", "peer", peer, "err", err)
	}

	delete(s.provisionedPeers, key)

	return nil
}

func (s *Switch) eventListener() {
	defer func() {
		if r := recover(); r != nil {
			s.Logger.Error("Panic in (*lp2p.Switch).eventListener", "panic", r)
		}
	}()

	s.Logger.Info("Starting event listener")

	for msg := range s.eventBusSubscription.Out() {
		switch tt := msg.(type) {
		case event.EvtPeerConnectednessChanged:
			if err := s.onPeerStatusChanged(tt); err != nil {
				s.Logger.Error(
					"Failed to handle onPeerConnectednessChanged",
					"err", err,
					"peer", tt.Peer.String(),
					"status", tt.Connectedness.String(),
				)
			}
		default:
			s.Logger.Error("Unknown event type skipped", "type", fmt.Sprintf("%T", tt))
		}
	}
}

func (s *Switch) onPeerStatusChanged(e event.EvtPeerConnectednessChanged) error {
	s.Logger.Info("Peer status update", "peer", e.Peer.String(), "status", e.Connectedness.String())

	// todo: do something with peerSet (eg mark as (dis)connected)

	return nil
}
