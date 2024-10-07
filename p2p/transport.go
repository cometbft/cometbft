package p2p

import (
	"net"

	na "github.com/cometbft/cometbft/p2p/netaddress"
	"github.com/cosmos/gogoproto/proto"
)

// peerConfig is used to bundle data we need to fully setup a Peer with an
// MConn, provided by the caller of Accept and Dial (currently the Switch). This
// a temporary measure until reactor setup is less dynamic and we introduce the
// concept of PeerBehaviour to communicate about significant Peer lifecycle
// events.
// TODO(xla): Refactor out with more static Reactor setup and PeerBehaviour.
type peerConfig struct {
	chDescs     []StreamDescriptor
	onPeerError func(Peer, any)
	outbound    bool
	// isPersistent allows you to set a function, which, given socket address
	// (for outbound peers) OR self-reported address (for inbound peers), tells
	// if the peer is persistent or not.
	isPersistent  func(*na.NetAddress) bool
	reactorsByCh  map[byte]Reactor
	msgTypeByChID map[byte]proto.Message
	metrics       *Metrics
}

// Transport emits and connects to Peers. The implementation of Peer is left to
// the transport. Each transport is also responsible to filter establishing
// peers specific to its domain.
type Transport interface {
	// Listening address.
	NetAddress() na.NetAddress

	// Accept returns a newly connected Peer.
	Accept() (net.Conn, na.NetAddress, error)

	// Dial connects to the Peer for the address.
	Dial(addr na.NetAddress) (net.Conn, error)

	// Cleanup any resources associated with Peer.
	Cleanup(peer Peer)
}
