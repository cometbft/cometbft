/*
 * Reactor is responsible for handling incoming messages on one or more
 * Channel. Switch calls GetChannels when reactor is added to it. When a new
 * peer joins our node, InitPeer and AddPeer are called. RemovePeer is called
 * when the peer is stopped. Receive is called when a message is received on a
 * channel associated with this reactor.
 */
// Code: https://github.com/cometbft/cometbft/blob/main/p2p/base_reactor.go
module reactor {

	/*
	 * Peer is an interface representing a peer connected on a reactor.
	 */
	// This is very minimal representation: only a string ID.
	type Peer = {
		ID: str,
	}

	// Peers that have been initialized but not yet removed.
	// The reactor should expect receiving messages from them.
	var peers: Set[Peer]

	// A Routine is used to interact with an active Peer.
	type Routine = {
		name: str,
		peerID: str,
	}

	// The reactor runs multiple routines.
	var routines: Set[Routine]

	// Get Routines associated to a Peer.
	def peerRoutines(peer: Peer): Set[Routine] = {
		routines.filter(r => r.peerID == peer.ID)
	}

	// Start routines to interact with a Peer.
	// Returns the started routines (can be multiple).
	def startRoutines(peer: Peer): Set[Routine] = {
		Set({name: "ioRoutine", peerID: peer.ID})
		//.union(Set({name: "auxRoutine", peerID: peer}))
	}

	// Stop routines interacting a Peer.
	// Returns the stopped routines (can be multiple).
	// FIXME: this method now does not do anything useful.
	def stopRoutines(peer: Peer): Set[Routine] = {
		peerRoutines(peer)
	}

	// Byte ID used by channels.
	type Byte = int

	// Channel configuration.
	// Only the Byte ID is relevant by now.
	type ChannelDescriptor = {
		ID: Byte,
		Priority: int,
	}

	/*
	 * GetChannels returns the list of MConnection.ChannelDescriptor. Make sure
	 * that each ID is unique across all the reactors added to the switch.
	 */
	// FIXME: could be a state variable, but it should never change.
	def GetChannels(): Set[ChannelDescriptor] = {
		Set({ID: 3, Priority: 1})
		//.union(Set({ID: 7, Priority: 2}))
	}

	// TODO: check that channel IDs are unique!

	/*
	 * InitPeer is called by the switch before the peer is started. Use it to
	 * initialize data for the peer (e.g. peer state).
	 */
	// TODO: returns a possibly updated peer.
	def InitPeer(peer: Peer): Peer = {
		peer
	}

	action initPeer(peer: Peer): bool = all {
		not(peer.in(peers)),
		val updatedPeer = InitPeer(peer)
		peers' = peers.union(Set(updatedPeer)),
		routines' = routines,
	}

	/* 
	 * AddPeer is called by the switch after the peer is added and successfully
	 * started. Use it to start goroutines communicating with the peer.
	 */
	action AddPeer(peer: Peer): bool = all {
		peer.in(peers), // initPeer(peer) and not RemovePeer(peer)
		peerRoutines(peer).size() == 0,	// StartPeer(peer) not called twice
		routines' = routines.union(startRoutines(peer)),
		peers' = peers,
	}

	/*
	 * RemovePeer is called by the switch when the peer is stopped (due to error
	 * or other reason).
	 */
	action RemovePeer(peer: Peer, reason: str): bool = all {
		peer.in(peers), // InitPeer(peer) and not RemovePeer(peer)
		// Routines could be already stopped if Peer has erroed.
		// Namely, we can have peerRoutines(peer).Size() == 0.
		routines' = routines.exclude(stopRoutines(peer)),
		peers' = peers.exclude(Set(peer)),
	}

	/*
	 * Envelope contains a message with sender routing info.
	 */
	type Envelope = {
		Src: Peer, // Sender
		Message: str, // Payload
		ChannelID: Byte,
	}

	/*
	 * Receive is called by the switch when an envelope is received from any connected
	 * peer on any of the channels registered by the reactor.
	 */
	// FIXME: action or def? In the current shape, it does not change any state.
	def Receive(e: Envelope): bool = all {
		e.Src.in(peers), // Sender is an active/connected peer
		GetChannels.exists(c => c.ID == e.ChannelID), // Registered for the channel
		// Process the payload: e.Message
	}

	// HELPERS

	action init = all {
		peers' = Set(),
		routines' = Set(),
	}

	def state() = {
		{peers: peers, routines: routines }
	}

//	// STATE of the P2P layer
//
//	// Code: var Switch.peers *PeerSet
//	// TODO: should ensure at most a Peer instance per ID
//	var peers: Set[Peer]
//
//	// ACTIONS taken by the P2P layer
//
//	// Code: Switch.addPeer(p Peer)
//	action peerConnected(p: Peer): bool = all {
//		not(p.in(peers)),	// Code: Switch.filterPeer(p Peer)
//		val peer = InitPeer(p)
//		all {
//			// TODO: peer.Start(), errors are failures
//			peers' = peers.union(Set(peer)),
//			AddPeer(peer),
//		}
//	}
//
//	// Code: Switch.stopAndRemovePeer(peer Peer, reason interface{})
//	action peerDisconnected(peer: Peer, reason: str): bool = all {
//		peer.in(peers),
//		// TODO: peer.Stop(), errors are only logged
//		RemovePeer(peer, reason),
//		peers' = peers.exclude(Set(peer))
//	}

}
