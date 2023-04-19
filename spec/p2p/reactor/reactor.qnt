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

	// Byte ID used by channels.
	type Byte = int

	// Channel configuration.
	// Only the Byte ID is relevant here.
	type ChannelDescriptor = {
		ID: Byte,
		Priority: int,
	}

	/*
	 * Envelope contains a message with sender routing info.
	 */
	type Envelope = {
		Src: Peer, // Sender
		Message: str, // Payload
		ChannelID: Byte,
	}

	// A Routine is used to interact with an active Peer.
	type Routine = {
		name: str,
		peerID: str,
	}

	type ReactorState = {
		// Peers that have been initialized but not yet removed.
		// The reactor should expect receiving messages from them.
		peers: Set[Peer],

		// The reactor runs multiple routines.
		routines: Set[Routine],

		// Service implementation.
		started: bool,
		stopped: bool,

		// Indicates whether a switch was set: reactor setup.
		switch: bool,
	}

	var state: ReactorState

	action init = all {
		state' = {
			peers: Set(),
			routines: Set(),
			started: false,
			stopped: false,
			switch: false,
		}
	}

	/*
	 * GetChannels returns the list of MConnection.ChannelDescriptor. Make sure
	 * that each ID is unique across all the reactors added to the switch.
	 */
	// TODO: we need a way to configure the returned channels!
	def GetChannels(s: ReactorState): Set[ChannelDescriptor] = {
		Set({ID: 3, Priority: 1})
		.union(Set({ID: 7, Priority: 2}))
		//.union(Set({ID: 3, Priority: 2})) // setup should fail!
	}

	/*
	 * SetSwitch allows setting a switch.
	 */
	def SetSwitch(s: ReactorState, switch: bool): ReactorState = {
		s.with("switch", switch)
	}

	// Code: Switch.AddReactor(name string, reactor Reactor)
	action setup() = all {
		not(state.switch),
		val channels = state.GetChannels()
		val chIDs = channels.map(c => c.ID)
		channels.size() == chIDs.size(), // channel IDs are unique
		state' = state.SetSwitch(true),
	}

	/*
	 * Start the service.
	 * If it's already started or stopped, will return an error.
	 */
	def OnStart(s: ReactorState): ReactorState = {
		s.with("started", true)
	}

	// Code: Switch.OnStart()
	action start(): bool = all {
		state.switch,
		not(state.started) and not(state.stopped),
		state' = state.OnStart()
	}

	/*
	 * Stop the service.
	 * If it's already stopped, will return an error.
	 */
	def OnStop(s: ReactorState): ReactorState = {
		s.with("stopped", true)
	}

	// Code: Switch.OnStop()
	action stop(): bool = all {
		state.started and not(state.stopped),
		state.peers.size() == 0, // RemovePeer(p) for all peers
		state' = state.OnStop()
	}

	/*
	 * Return true if the service is running.
	 */
	def IsRunning(s: ReactorState): bool = {
		s.started and not(s.stopped)
	}

	/*
	 * InitPeer is called by the switch before the peer is started. Use it to
	 * initialize data for the peer (e.g. peer state).
	 */
	// TODO: the interface method should return an updated peer!
	def InitPeer(s: ReactorState, peer: Peer): ReactorState = {
		s.with("peers", s.peers.union(Set(peer)))
	}

	// Code: Switch.addPeer(p Peer): preamble
	action initPeer(peer: Peer): bool = all {
		state.switch and state.IsRunning(),
		not(peer.in(state.peers)), // No previous InitPeer(peer)
		state' = state.InitPeer(peer)
	}

	/* 
	 * AddPeer is called by the switch after the peer is added and successfully
	 * started. Use it to start goroutines communicating with the peer.
	 */
	def AddPeer(s: ReactorState, peer: Peer): ReactorState = {
		val startedRoutines = Set( {name: "ioRoutine", peerID: peer.ID} )
		s.with("routines", s.routines.union(startedRoutines))
	}

	// Code: Switch.addPeer(p Peer): conclusion
	action addPeer(peer: Peer): bool = all {
		state.switch and state.IsRunning(),
		peer.in(state.peers), // InitPeer(peer) and not RemovePeer(peer)
		state.routines.filter(r => r.peerID == peer.ID).size() == 0,
		state' = AddPeer(state, peer)
	}

	/*
	 * RemovePeer is called by the switch when the peer is stopped (due to error
	 * or other reason).
	 */
	def RemovePeer(s: ReactorState, peer: Peer, reason: str): ReactorState = {
		val stoppedRoutines = state.routines.filter(r => r.peerID == peer.ID)
		s.with("peers", s.peers.exclude(Set(peer)))
		 .with("routines", s.routines.exclude(stoppedRoutines))
	}

	// Code: Switch.stopAndRemovePeer(peer Peer, reason interface{})
	action removePeer(peer: Peer, reason: str): bool = all {
		state.switch and state.IsRunning(),
		peer.in(state.peers), // InitPeer(peer) and not RemovePeer(peer)
		// Routines might not be started, namely: not AddPeer(peer)
		// Routines could also be already stopped if Peer has erroed.
		state' = state.RemovePeer(peer, reason)
	}

	/*
	 * Receive is called by the switch when an envelope is received from any connected
	 * peer on any of the channels registered by the reactor.
	 */
	def Receive(s: ReactorState, e: Envelope): ReactorState = {
		// Process the payload: e.Message
		s
	}

	action receive(e: Envelope): bool = all {
		state.switch and state.IsRunning(),
		e.Src.in(state.peers), // Sender is an active/connected peer
		state.GetChannels().exists(
			c => c.ID == e.ChannelID), // Registered for the channel
		state' = Receive(state, e)
	}

}
