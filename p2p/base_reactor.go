package p2p

import (
	"github.com/cometbft/cometbft/libs/service"
	"github.com/cometbft/cometbft/p2p/conn"
)

// Reactor is responsible for handling incoming messages on one or more
// Channel. Switch calls GetChannels when reactor is added to it. When a new
// peer joins our node, InitPeer and AddPeer are called. RemovePeer is called
// when the peer is stopped. Receive is called when a message is received on a
// channel associated with this reactor.
//
// Peer#Send or Peer#TrySend should be used to send the message to a peer.
type Reactor interface {
	service.Service // Start, Stop

	// SetSwitch allows setting a switch.
	SetSwitch(Switcher)

	// GetChannels returns the list of MConnection.ChannelDescriptor. Make sure
	// that each ID is unique across all the reactors added to the switch.
	GetChannels() []*conn.ChannelDescriptor

	// InitPeer is called by the switch before the peer is started. Use it to
	// initialize data for the peer (e.g. peer state).
	//
	// NOTE: If the peer is not added after InitPeer (it fails to start, or the
	// peer set rejects it), the switch calls RemovePeer without calling
	// AddPeer, and the peer may not have been stopped yet. The exception is a
	// rejected duplicate of a peer ID that is already connected: RemovePeer is
	// not called then, because it would clear the state of the connected peer.
	// InitPeer must therefore be idempotent for a given peer ID, and data
	// stored in the reactor itself must be cleaned up in RemovePeer.
	InitPeer(peer Peer) Peer

	// AddPeer is called by the switch after the peer is added and successfully
	// started. Use it to start goroutines communicating with the peer.
	AddPeer(peer Peer)

	// RemovePeer is called by the switch when the peer is stopped (due to error
	// or other reason), or when the switch fails to add a peer after InitPeer
	// was called for it.
	RemovePeer(peer Peer, reason any)

	// Receive is called by the switch when an envelope is received from any connected
	// peer on any of the channels registered by the reactor
	Receive(Envelope)
}

// MsgBytesFilter is an optional interface a Reactor may implement to inspect
// raw message bytes on a channel before they are protobuf unmarshalled. If
// FilterMsgBytes returns a non-nil error, the message is dropped and the peer
// is disconnected.
type MsgBytesFilter interface {
	FilterMsgBytes(chID byte, src Peer, msgBytes []byte) error
}

//--------------------------------------

type BaseReactor struct {
	service.BaseService // Provides Start, Stop, .Quit
	Switch              Switcher
}

func NewBaseReactor(name string, impl Reactor) *BaseReactor {
	return &BaseReactor{
		BaseService: *service.NewBaseService(nil, name, impl),
		Switch:      nil,
	}
}

func (br *BaseReactor) SetSwitch(sw Switcher) {
	br.Switch = sw
}
func (*BaseReactor) GetChannels() []*conn.ChannelDescriptor { return nil }
func (*BaseReactor) AddPeer(Peer)                           {}
func (*BaseReactor) RemovePeer(Peer, any)                   {}
func (*BaseReactor) Receive(Envelope)                       {}
func (*BaseReactor) InitPeer(peer Peer) Peer                { return peer }
