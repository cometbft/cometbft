package p2p

// Switcher handles peer connections and exposes an API to receive incoming messages
// on `Reactors`. Each `Reactor` is responsible for handling incoming messages of one
// or more `Channels`. So while sending outgoing messages is typically performed on the peer,
// incoming messages are received on the reactor.
type Switcher interface {
	Reactor(name string) (reactor Reactor, exists bool)

	Broadcast(e Envelope) (successChan chan bool)

	Peers() IPeerSet
	NumPeers() (outbound, inbound, dialing int)
	MaxNumOutboundPeers() int

	DialPeerWithAddress(addr *NetAddress) error

	StopPeerForError(peer Peer, reason any)
	StopPeerGracefully(peer Peer)

	MarkPeerAsGood(peer Peer)

	IsDialingOrExistingAddress(addr *NetAddress) bool
	IsPeerPersistent(addr *NetAddress) bool
	IsPeerUnconditional(id ID) bool
}

var _ Switcher = (*Switch)(nil)
