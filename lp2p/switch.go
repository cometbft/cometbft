package lp2p

import (
	"fmt"

	"github.com/cometbft/cometbft/config"
	"github.com/cometbft/cometbft/libs/log"
	"github.com/cometbft/cometbft/libs/service"
	"github.com/cometbft/cometbft/p2p"
	"github.com/libp2p/go-libp2p/core/network"
)

// Switch represents p2p.Switcher alternative implementation based on go-libp2p.
type Switch struct {
	service.BaseService

	config   *config.P2PConfig
	nodeInfo p2p.NodeInfo // our node info
	nodeKey  *p2p.NodeKey // our node private key

	host *Host

	reactors map[string]p2p.Reactor
}

// ReactorItem is a pair of name and reactor.
// Preserves order when adding.
type ReactorItem struct {
	Name    string
	Reactor p2p.Reactor
}

var _ p2p.Switcher = (*Switch)(nil)

// NewSwitch constructs a new Switch.
func NewSwitch(
	cfg *config.P2PConfig,
	nodeInfo p2p.NodeInfo,
	nodeKey *p2p.NodeKey,
	host *Host,
	reactors []ReactorItem,
	logger log.Logger,
) *Switch {
	s := &Switch{
		config:   cfg,
		reactors: make(map[string]p2p.Reactor),
		nodeInfo: nodeInfo,
		nodeKey:  nodeKey,
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
	s.logUnimplemented("RemoveReactor")
}

// --------------------------------
// PeerManager methods
// --------------------------------

func (s *Switch) Peers() p2p.IPeerSet {
	panic("unimplemented")
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

	// note we don't count dialing peers here

	return outbound, inbound, dialing
}

func (s *Switch) MaxNumOutboundPeers() int {
	s.logUnimplemented("MaxNumOutboundPeers")

	return 0
}

func (s *Switch) AddPersistentPeers(addrs []string) error {
	panic("unimplemented")
}

func (s *Switch) AddPrivatePeerIDs(ids []string) error {
	panic("unimplemented")
}

func (s *Switch) AddUnconditionalPeerIDs(ids []string) error {
	panic("unimplemented")
}

func (s *Switch) SetAddrBook(addrBook p2p.AddrBook) {
	panic("unimplemented")
}

func (s *Switch) DialPeerWithAddress(_ *p2p.NetAddress) error {
	s.logUnimplemented("DialPeerWithAddress")

	return nil
}

func (s *Switch) DialPeersAsync(peers []string) error {
	panic("unimplemented")
}

func (s *Switch) StopPeerGracefully(_ p2p.Peer) {
	s.logUnimplemented("StopPeerGracefully")
}

func (s *Switch) StopPeerForError(peer p2p.Peer, reason any) {
	panic("unimplemented")
}

func (s *Switch) IsDialingOrExistingAddress(addr *p2p.NetAddress) bool {
	panic("unimplemented")
}

func (s *Switch) IsPeerPersistent(addr *p2p.NetAddress) bool {
	panic("unimplemented")
}

func (s *Switch) IsPeerUnconditional(id p2p.ID) bool {
	panic("unimplemented")
}

func (s *Switch) MarkPeerAsGood(_ p2p.Peer) {
	s.logUnimplemented("MarkPeerAsGood")
}

//--------------------------------
// Broadcaster methods
//--------------------------------

func (s *Switch) Broadcast(e p2p.Envelope) (successChan chan bool) {
	panic("unimplemented")
}

func (s *Switch) BroadcastAsync(e p2p.Envelope) {
	panic("unimplemented")
}

func (s *Switch) TryBroadcast(e p2p.Envelope) {
	panic("unimplemented")
}

func (s *Switch) logUnimplemented(method string) {
	s.Logger.Info("Unimplemented LibP2PSwitch method called", "method", method)
}
