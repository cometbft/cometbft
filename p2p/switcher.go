package p2p

import (
	"github.com/cometbft/cometbft/libs/log"
	"github.com/cometbft/cometbft/libs/service"
)

// Switcher handles peer connections and exposes an API to receive incoming messages
// on `Reactors`. Each `Reactor` is responsible for handling incoming messages of one
// or more `Channels`. So while sending outgoing messages is typically performed on the peer,
// incoming messages are received on the reactor.
type Switcher interface {
	service.Service

	ReactorManager
	PeerManager
	Broadcaster

	NodeInfo() NodeInfo
	Log() log.Logger
}

type ReactorManager interface {
	Reactor(name string) (reactor Reactor, exists bool)
	AddReactor(name string, reactor Reactor) Reactor
	RemoveReactor(name string, reactor Reactor)
}

type PeerManager interface {
	Peers() IPeerSet
	NumPeers() (outbound, inbound, dialing int)
	MaxNumOutboundPeers() int

	AddPersistentPeers(addrs []string) error
	AddPrivatePeerIDs(ids []string) error
	AddUnconditionalPeerIDs(ids []string) error

	DialPeerWithAddress(addr *NetAddress) error
	DialPeersAsync(peers []string) error

	StopPeerForError(peer Peer, reason any)
	StopPeerGracefully(peer Peer)

	IsDialingOrExistingAddress(addr *NetAddress) bool
	IsPeerPersistent(addr *NetAddress) bool
	IsPeerUnconditional(id ID) bool

	MarkPeerAsGood(peer Peer)
}

type Broadcaster interface {
	Broadcast(e Envelope) (successChan chan bool)
	BroadcastAsync(e Envelope)
	TryBroadcast(e Envelope)
}

var _ Switcher = (*Switch)(nil)
