package p2p

import (
	"fmt"

	"github.com/cometbft/cometbft/config"
	"github.com/cometbft/cometbft/libs/log"
	"github.com/cometbft/cometbft/libs/service"
	"github.com/cometbft/cometbft/p2p/lp2p"
	"github.com/libp2p/go-libp2p/core/network"
)

// LibP2PSwitch represents alternative Switcher implementation based on go-libp2p.
type LibP2PSwitch struct {
	service.BaseService

	config   *config.P2PConfig
	nodeInfo NodeInfo // our node info
	nodeKey  *NodeKey // our node private key

	host *lp2p.Host

	reactors map[string]Reactor
}

// ReactorItem is a pair of name and reactor.
// Preserves order when adding.
type ReactorItem struct {
	Name    string
	Reactor Reactor
}

var _ Switcher = (*LibP2PSwitch)(nil)

// NewSwitchLibP2P constructs a new SwitchLibP2P.
func NewLibP2PSwitch(
	cfg *config.P2PConfig,
	nodeInfo NodeInfo,
	nodeKey *NodeKey,
	host *lp2p.Host,
	reactors []ReactorItem,
	logger log.Logger,
) *LibP2PSwitch {
	s := &LibP2PSwitch{
		config:   cfg,
		reactors: make(map[string]Reactor),
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

func (s *LibP2PSwitch) OnStart() error {
	s.Logger.Info("Starting LibP2PSwitch")

	for name, reactor := range s.reactors {
		err := reactor.OnStart()
		if err != nil {
			return fmt.Errorf("failed to start reactor %s: %w", name, err)
		}
	}

	return nil
}

func (s *LibP2PSwitch) OnStop() {
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

func (s *LibP2PSwitch) NodeInfo() NodeInfo {
	return s.nodeInfo
}

func (s *LibP2PSwitch) Log() log.Logger {
	return s.Logger
}

//--------------------------------
// ReactorManager methods
//--------------------------------

func (s *LibP2PSwitch) Reactor(name string) (Reactor, bool) {
	reactor, exists := s.reactors[name]

	return reactor, exists
}

func (s *LibP2PSwitch) AddReactor(name string, reactor Reactor) Reactor {
	// todo register channels !!!

	s.reactors[name] = reactor
	reactor.SetSwitch(s)

	return reactor
}

func (s *LibP2PSwitch) RemoveReactor(_ string, _ Reactor) {
	s.logUnimplemented("RemoveReactor")
}

// --------------------------------
// PeerManager methods
// --------------------------------

func (s *LibP2PSwitch) Peers() IPeerSet {
	panic("unimplemented")
}

func (s *LibP2PSwitch) NumPeers() (outbound, inbound, dialing int) {
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

func (s *LibP2PSwitch) MaxNumOutboundPeers() int {
	s.logUnimplemented("MaxNumOutboundPeers")

	return 0
}

func (s *LibP2PSwitch) AddPersistentPeers(addrs []string) error {
	panic("unimplemented")
}

func (s *LibP2PSwitch) AddPrivatePeerIDs(ids []string) error {
	panic("unimplemented")
}

func (s *LibP2PSwitch) AddUnconditionalPeerIDs(ids []string) error {
	panic("unimplemented")
}

func (s *LibP2PSwitch) SetAddrBook(addrBook AddrBook) {
	panic("unimplemented")
}

func (s *LibP2PSwitch) DialPeerWithAddress(_ *NetAddress) error {
	s.logUnimplemented("DialPeerWithAddress")

	return nil
}

func (s *LibP2PSwitch) DialPeersAsync(peers []string) error {
	panic("unimplemented")
}

func (s *LibP2PSwitch) StopPeerGracefully(_ Peer) {
	s.logUnimplemented("StopPeerGracefully")
}

func (s *LibP2PSwitch) StopPeerForError(peer Peer, reason any) {
	panic("unimplemented")
}

func (s *LibP2PSwitch) IsDialingOrExistingAddress(addr *NetAddress) bool {
	panic("unimplemented")
}

func (s *LibP2PSwitch) IsPeerPersistent(addr *NetAddress) bool {
	panic("unimplemented")
}

func (s *LibP2PSwitch) IsPeerUnconditional(id ID) bool {
	panic("unimplemented")
}

func (s *LibP2PSwitch) MarkPeerAsGood(_ Peer) {
	s.logUnimplemented("MarkPeerAsGood")
}

//--------------------------------
// Broadcaster methods
//--------------------------------

func (s *LibP2PSwitch) Broadcast(e Envelope) (successChan chan bool) {
	panic("unimplemented")
}

func (s *LibP2PSwitch) BroadcastAsync(e Envelope) {
	panic("unimplemented")
}

func (s *LibP2PSwitch) TryBroadcast(e Envelope) {
	panic("unimplemented")
}

func (s *LibP2PSwitch) logUnimplemented(method string) {
	s.Logger.Info("Unimplemented LibP2PSwitch method called", "method", method)
}
