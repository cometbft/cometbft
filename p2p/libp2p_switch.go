package p2p

import (
	"github.com/cometbft/cometbft/config"
	"github.com/cometbft/cometbft/libs/log"
	"github.com/cometbft/cometbft/libs/service"
)

// LibP2PSwitch represents alternative Switcher implementation based on go-libp2p.
type LibP2PSwitch struct {
	service.BaseService

	config   *config.P2PConfig
	nodeInfo NodeInfo // our node info
	nodeKey  *NodeKey // our node private key

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

func (s *LibP2PSwitch) Start() error {
	panic("unimplemented")
}

func (s *LibP2PSwitch) Stop() error {
	panic("unimplemented")
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

func (s *LibP2PSwitch) Reactor(name string) (reactor Reactor, exists bool) {
	panic("unimplemented")
}

func (s *LibP2PSwitch) AddReactor(name string, reactor Reactor) Reactor {
	panic("unimplemented")
}

func (s *LibP2PSwitch) RemoveReactor(name string, reactor Reactor) {
	panic("unimplemented")
}

// --------------------------------
// PeerManager methods
// --------------------------------

func (s *LibP2PSwitch) Peers() IPeerSet {
	panic("unimplemented")
}

func (s *LibP2PSwitch) NumPeers() (outbound int, inbound int, dialing int) {
	panic("unimplemented")
}

func (s *LibP2PSwitch) MaxNumOutboundPeers() int {
	panic("unimplemented")
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

func (s *LibP2PSwitch) DialPeerWithAddress(addr *NetAddress) error {
	panic("unimplemented")
}

func (s *LibP2PSwitch) DialPeersAsync(peers []string) error {
	panic("unimplemented")
}

func (s *LibP2PSwitch) StopPeerGracefully(peer Peer) {
	panic("unimplemented")
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

func (s *LibP2PSwitch) MarkPeerAsGood(peer Peer) {
	panic("unimplemented")
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
