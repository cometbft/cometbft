package p2p

import (
	"net"

	"github.com/cosmos/gogoproto/proto"

	na "github.com/cometbft/cometbft/p2p/netaddress"
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
	// NetAddress returns the NetAddress of the local node.
	NetAddress() na.NetAddress

	// Accept waits for and returns the next connection to the local node.
	Accept() (net.Conn, *na.NetAddress, error)

	// Dial dials the given address and returns a connection.
	Dial(addr na.NetAddress) (net.Conn, error)

	// Cleanup any resources associated with the given connection.
	//
	// Must be run when the peer is dropped for any reason.
	Cleanup(conn net.Conn) error
}
