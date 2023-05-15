// -*- mode: Bluespec; -*-
/*
 * Reactor is responsible for handling incoming messages on one or more
 * Channel. Switch calls GetChannels when reactor is added to it. When a new
 * peer joins our node, InitPeer and AddPeer are called. RemovePeer is called
 * when the peer is stopped. Receive is called when a message is received on a
 * channel associated with this reactor.
 */
// Code: https://github.com/cometbft/cometbft/blob/main/p2p/base_reactor.go
module reactor {

	// Unique ID of a node.
	type NodeID = str

	/*
	 * Peer is an interface representing a peer connected on a reactor.
	 */
	type Peer = {
		ID: NodeID,

		// Other fields can be added to represent the p2p operation.
	}

	// Byte ID used by channels, must be globally unique.
	type Byte = str

	// Channel configuration.
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
		peer: Peer,
	}

	type ReactorState = {
		// Peers that have been initialized but not yet removed.
		// The reactor should expect receiving messages from them.
		peers: Set[NodeID],

		// The reactor runs multiple routines.
		routines: Set[Routine],

		// Service implementation.
		started: bool,
		stopped: bool,

		// Indicates whether the reactor was registered.
		registered: bool,
	}

	// Produces a new, uninitialized reactor.
	pure def NewReactor(): ReactorState = {
		{
			peers: Set(),
			routines: Set(),
			started: false,
			stopped: false,
			registered: false,
		}
	}

	// Pure definitions below represent the `p2p.Reactor` interface methods:

	/*
	 * GetChannels returns the list of MConnection.ChannelDescriptor. Make sure
	 * that each ID is unique across all the reactors added to the switch.
	 */
	pure def GetChannels(s: ReactorState): Set[ChannelDescriptor] = {
		// This method should return a list of channels, e.g.:
		Set({ID: "3", Priority: 1}, {ID: "7", Priority: 2})
	}

	/*
	 * SetSwitch allows setting a switch.
	 */
	pure def SetSwitch(s: ReactorState, switch: bool): ReactorState = {
		s.with("registered", switch)
	}

	/*
	 * Start the service.
	 * If it's already started or stopped, will return an error.
	 */
	pure def OnStart(s: ReactorState): ReactorState = {
		// Startup procedures should come here.
		s.with("started", true)
	}

	/*
	 * Stop the service.
	 * If it's already stopped, will return an error.
	 */
	pure def OnStop(s: ReactorState): ReactorState = {
		// Shutdown procedures should come here.
		s.with("stopped", true)
	}

	/*
	 * InitPeer is called by the switch before the peer is started. Use it to
	 * initialize data for the peer (e.g. peer state).
	 */
	pure def InitPeer(s: ReactorState, peer: Peer): (ReactorState, Peer) = {
		// This method can update the received peer, which is returned.
		val updatedPeer = peer
		(s.with("peers", s.peers.union(Set(peer.ID))), updatedPeer)
	}

	/* 
	 * AddPeer is called by the switch after the peer is added and successfully
	 * started. Use it to start goroutines communicating with the peer.
	 */
	pure def AddPeer(s: ReactorState, peer: Peer): ReactorState = {
		// This method can be used to start routines to handle the peer.
		// Below an example of an arbitrary 'ioRoutine' routine.
		val startedRoutines = Set( {name: "ioRoutine", peer: peer} )
		s.with("routines", s.routines.union(startedRoutines))
	}

	/*
	 * RemovePeer is called by the switch when the peer is stopped (due to error
	 * or other reason).
	 */
	pure def RemovePeer(s: ReactorState, peer: Peer, reason: str): ReactorState = {
		// This method should stop routines created by `AddPeer(Peer)`.
		val stoppedRoutines = s.routines.filter(r => r.peer.ID == peer.ID)
		s.with("peers", s.peers.exclude(Set(peer.ID)))
		 .with("routines", s.routines.exclude(stoppedRoutines))
	}

	/*
	 * Receive is called by the switch when an envelope is received from any connected
	 * peer on any of the channels registered by the reactor.
	 */
	pure def Receive(s: ReactorState, e: Envelope): ReactorState = {
		// This method should process the message payload: e.Message.
		s
	}

	// Global state

	var state: ReactorState

	pure def IsRunning(s: ReactorState): bool = {
		s.started and not(s.stopped)
	}

	// State transitions performed by the p2p layer, invoking `p2p.Reactor` methods:

	// Code: Switch.AddReactor(name string, reactor Reactor)
	action register() = all {
		not(state.registered),
		val channels = state.GetChannels()
		val chIDs = channels.map(c => c.ID)
		channels.size() == chIDs.size(), // channel IDs are unique
		state' = state.SetSwitch(true)
	}

	// Code: Switch.OnStart()
	action start() = all {
		state.registered,
		not(state.started) and not(state.stopped),
		state' = state.OnStart()
	}

	// Code: Switch.addPeer(p Peer): preamble
	action initPeer(peer: Peer): bool = all {
		state.registered and state.IsRunning(),
		not(peer.ID.in(state.peers)),
		state' = state.InitPeer(peer)._1 // Updated peer not used here
	}

	// Code: Switch.addPeer(p Peer): conclusion
	action addPeer(peer: Peer): bool = all {
		state.registered and state.IsRunning(),
		peer.ID.in(state.peers), // InitPeer(peer) and not RemovePeer(peer)
		state.routines.filter(r => r.peer.ID == peer.ID).size() == 0,
		state' = AddPeer(state, peer)
	}

	// Code: Switch.stopAndRemovePeer(peer Peer, reason interface{})
	action removePeer(peer: Peer, reason: str): bool = all {
		state.registered and state.IsRunning(),
		peer.ID.in(state.peers), // InitPeer(peer) and not RemovePeer(peer)
		// Routines might not be started, namely: not AddPeer(peer)
		// Routines could also be already stopped if Peer has erroed.
		state' = state.RemovePeer(peer, reason)
	}

	// Code: Peer type, onReceive := func(chID byte, msgBytes []byte)
	action receive(e: Envelope): bool = all {
		state.registered and state.IsRunning(),
		e.Src.ID.in(state.peers), // Sender is an active peer
		state.GetChannels().exists(
			c => c.ID == e.ChannelID), // Registered for the channel
		state' = Receive(state, e)
	}

	// Code: Switch.OnStop()
	action stop = all {
		state.registered and IsRunning(state),
		// Either no peer was added or all peers were removed
		state.peers.size() == 0,
		state' = state.OnStop()
	}

	// Simulation support

	action init = all {
		state' = NewReactor
	}

	pure val samplePeers = Set({ID: "p1"}, {ID: "p3"})
	pure val sampleChIDs = Set("1", "3", "7") // ChannelID 1 not registered
	pure val sampleMsgs  = Set("ping", "pong")

	action step = any {
		// One-shot actions
		register,
		start,
		stop,
		// Peer-specific actions
		nondet peer = oneOf(samplePeers)
		any {
			initPeer(peer),
			addPeer(peer),
			removePeer(peer, "no reason"),
			receive({Src: peer,
				ChannelID: oneOf(sampleChIDs),
				Message: oneOf(sampleMsgs)}),
		}
	}

}
