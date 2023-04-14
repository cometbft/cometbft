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
	type Peer = str

	/*
	 * InitPeer is called by the switch before the peer is started. Use it to
	 * initialize data for the peer (e.g. peer state).
	 */
	// Returns a possibly updated peer.
	def InitPeer(peer: Peer): Peer = {
		peer
	}

	/* 
	 * AddPeer is called by the switch after the peer is added and successfully
	 * started. Use it to start goroutines communicating with the peer.
	 */
	// Does not return anything.
	def AddPeer(peer: Peer): bool = {
		true
	}

	/*
	 * RemovePeer is called by the switch when the peer is stopped (due to error
	 * or other reason).
	 */
	// Does not return anything.
	def RemovePeer(peer: Peer, reason: str): bool = {
		true
	}

	// STATE of the P2P layer

	// Code: var Switch.peers *PeerSet
	// TODO: should ensure at most a Peer instance per ID
	var peers: Set[Peer]

	action init = all {
		peers' = Set(),
	}

	// ACTIONS taken by the P2P layer

	// Code: Switch.addPeer(p Peer)
	action peerConnected(p: Peer): bool = all {
		not(p.in(peers)),	// Code: Switch.filterPeer(p Peer)
		val peer = InitPeer(p)
		all {
			// TODO: peer.Start(), errors are failures
			peers' = peers.union(Set(peer)),
			AddPeer(peer),
		}
	}

	// Code: Switch.stopAndRemovePeer(peer Peer, reason interface{})
	action peerDisconnected(peer: Peer, reason: str): bool = all {
		peer.in(peers),
		// TODO: peer.Stop(), errors are only logged
		RemovePeer(peer, reason),
		peers' = peers.exclude(Set(peer))
	}

}
