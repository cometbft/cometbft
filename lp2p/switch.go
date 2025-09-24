package lp2p

import (
	"fmt"

	"github.com/cometbft/cometbft/config"
	"github.com/cometbft/cometbft/libs/log"
	"github.com/cometbft/cometbft/libs/service"
	"github.com/cometbft/cometbft/p2p"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/pkg/errors"
)

// Switch represents p2p.Switcher alternative implementation based on go-libp2p.
type Switch struct {
	service.BaseService

	config   *config.P2PConfig
	nodeKey  *p2p.NodeKey // our node private key
	nodeInfo p2p.NodeInfo // our node info

	host    *Host
	peerSet *PeerSet

	reactors map[string]p2p.Reactor
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

		reactors: make(map[string]p2p.Reactor),
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

	for name, reactor := range s.reactors {
		err := reactor.OnStart()
		if err != nil {
			return fmt.Errorf("failed to start reactor %s: %w", name, err)
		}
	}

	return nil
}

func (s *Switch) OnStop() {
	s.Logger.Info("Stopping LibP2PSwitch")

	for name, reactor := range s.reactors {
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
	reactor, exists := s.reactors[name]

	return reactor, exists
}

func (s *Switch) AddReactor(name string, reactor p2p.Reactor) p2p.Reactor {
	// todo register channels !!!

	s.reactors[name] = reactor
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

	if err := s.host.Network().ClosePeer(p.addrInfo.ID); err != nil {
		s.Logger.Error("Failed to close peer", "peer", peer, "err", err)
	}

	s.peerSet.Remove(peer.ID())
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

func (s *Switch) Broadcast(e p2p.Envelope) (successChan chan bool) {
	// todo
	panic("unimplemented")
}

func (s *Switch) BroadcastAsync(e p2p.Envelope) {
	// todo
	panic("unimplemented")
}

func (s *Switch) TryBroadcast(e p2p.Envelope) {
	// todo
	panic("unimplemented")
}

func (s *Switch) logUnimplemented(method string, kv ...any) {
	s.Logger.Info(
		"Unimplemented LibP2PSwitch method called",
		append(kv, "method", method)...,
	)
}
