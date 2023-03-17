//
// S T O P   R E A D I N G ! ! ! This does not work, does not compile. I am in the process of refactoring for the multi process setting:
//

/*
  This encodes the execution flows of the p2p system of CometBFT for maintaining connections to other
  nodes in the network. It is based on the descriptions in 
  https://github.com/cometbft/cometbft/blob/91f8dc3ea0490fd824d1d3d346fb0da61de2f848/spec/p2p/architecture/peers-connect.md?plain=1#L1

  The funcitionality described here is reactive. It is triggered by
  - The PEX reactor (actually a part of the p2p system)
  - CometBFT reactors
  - connection requests from other nodes in the network
  - node operators
  - lower level code (TCP connection management)

  Josef Widder, Informal Systems, 2023
*/

module p2p {

    // TODO: "pure val" is a work around as constants are not working in the REPL, yet
    pure val peerStates = Set("dialing", "reconnecting", "inbound", "outbound")
    pure val connectionStates = Set("init", "open", "closed")
    // TODO: We might experiment around the max peer numbers. Should we model them as variables?
    pure val MaxNumOutboundPeers = 5
    pure val MaxNumInboundPeers = 10

    // (public_key, ipAddress:port)
    type Address = {ID:str, networkAddress:str}

    // p2p variables 

//var peerState: Address -> Address -> str
 //   var connectionState: Address -> Address -> str
 //   var IsUnconditional: Address -> Set[Address]

    // PEX variable
    //var AddressBook: Address -> Set[Address]
    // TODO: should be constant (set via instances). Doesn't work, yet.
    //var PersistentPeers: Address -> Set[Address] // typically outbound.

    // Environment. These are the nodes that may dial us
    // TODO: should be constant (set via instances). Doesn't work, yet.
    var Environment: Set[Address]
 
    // Operator. The operator can manually provide a list of addresses to dial via dialPeers
    // TODO: should be constant (set via instances). Doesn't work, yet.
    var OperatorAddressBook: Set[Address]

    //    EnvironmentX: Set[Address],
    //    OperatorAddressBook: Set[Address],

    type LocalState = {
        peerState: Address -> str,
        connectionState: Address -> str,
        IsUnconditional: Set[Address],
        AddressBook: Set[Address],
        PersistentPeers: Set[Address],
    }

    def Peers(v: LocalState): Set[Address] = v.peerState.keys()

    var configuration: Address -> LocalState

    // TODO: write functions as definitions.
    // Then write actions that use the definitions, and map "local" defintions as they act on the global statte
    // Write actions that describe the environment

    def dialPeer(s: LocalState, a: Address): LocalState = {
        if (s.peersOutbound() + s.peersDialing() < MaxNumOutboundPeers and not(s.Peers())) {
            s.with("peerState", s.peerState.set(a, "dialing"))
                .with("connectionState", s.connectionState.set(a, "init"))
        } else 
            s
    }

    action actionDialPeer (node: Address):bool {
        peerViews' = peerViews.set(node, dialPeer(peerViews.get(node)))

    }

    // What we call Peers is defined the domain of the peerState and connectionState maps. 
    // They should be the same, so we can check the invariant invPeers

    def explicitNodes = peerState.keys()
    // auxiliary definition. These are the peers that seem alive and connected
    def activePeers(p) = Peers(p).filter(x => connectionState.get(p).get(x) == "open")

    val invPeers = explicitNodes.forall(x => peerState.get(x).keys() == connectionState.get(x).keys())

    val invConnectionsNotOpen = explicitNodes.forall(y => Peers(y).forall(x => 
        (peerState.get(y).get(x) == "dialing" or peerState.get(y).get(x) == "reconnecting") implies connectionState.get(y).get(x) != "open"))

    val invConnectionsNotInit = explicitNodes.forall(y => Peers(y).forall(x => 
        (peerState.get(y).get(x) == "inbound" or peerState.get(y).get(x) == "outbound") implies connectionState.get(y).get(x) != "init"))

    val invConnectionsOpen = explicitNodes.forall(y => activePeers(y).forall(x => peerState.get(y).get(x) == "outbound" or peerState.get(y).get(x) == "inbound"))

    // counting peers in different states
    def peersOutbound(p) = size(Peers(p).filter(a => peerState.get(p).get(a) == "outbound"))
    def peersInbound(p) = size(Peers(p).filter(a => peerState.get(p).get(a) == "inbound"))
    def peersDialing(p) = size(Peers(p).filter(a => peerState.get(p).get(a) == "dialing"))

    // auxiliary definition for PEX, 
    // If we don't have peers in the addressbook to try, we dial Seeds to ask for new ones.
    def allAddressesArePeers(p) = AddressBook.get(p).exclude(Peers(p)) == Set()

    // helps for better readability. Future syntactic sugar 
    def require(pred:bool):bool = pred

    // This is painful
    action allUnchanged = all {
        peerState' = peerState,
        connectionState' = connectionState,
        AddressBook' = AddressBook,
        Environment' = Environment,
        IsUnconditional' = IsUnconditional,
        PersistentPeers' = PersistentPeers,
        OperatorAddressBook' = OperatorAddressBook,

    }    

    /*
    We will distinguish three kinds of actions according to their origin
    - low level by "the network". This includes incoming requests, connections being lost, but also
      the result of low-level protocol executions (e.g., setting a connection to connected after it was 
      succesful established)
    - reactors: reactors can ask p2p to remove peers if they fail w.r.t. reactor protocol logic
    - PEX reactor: This is a specific reactor that makes sure "enough" connections are established.
    */


    def setConnectionState(node, a, newval) = 
        connectionState.setBy(node, (old => old.set(a, newval)))

    def setPeerState(node, a, newval) = 
        peerState.setBy(node, (old => old.set(a, newval)))
 
    def putConnectionState(node, a, newval) = 
        connectionState.setBy(node, (old => old.put(a, newval)))

    def putPeerState(node, a, newval) = 
        peerState.setBy(node, (old => old.put(a, newval)))
    

    // triggered by "the network"
    action loseConnection(node: Address, a:Address):bool = all {
        require(connectionState.get(node).get(a) != "closed"),
        // The following line is obscure magic and we discovered a bug in the cheat sheet
        connectionState' = setConnectionState(node, a, "closed"),
        all { //unchanged
            AddressBook' = AddressBook,
            peerState' = peerState,
            Environment' = Environment,
            IsUnconditional' = IsUnconditional,
            PersistentPeers' = PersistentPeers,
            OperatorAddressBook' = OperatorAddressBook,
        }
    }

    // TODO: not sure what triggers that in the code.
    // We try to reconnect to an existing peer, otherwise we just clean up
    action onLostConnection(node: Address, a:Address):bool = all {
        require(connectionState.get(node).get(a) == "closed"),
        if (PersistentPeers.get(node).contains(a)) all {
            peerState' = setPeerState(node, a, "reconnecting"),
            connectionState' = setConnectionState(node, a, "dialing"),
        }
        else all {
            // if peer is not persistent, we just delete it
            // it may still be in some addressbook. Thus upon request (e.g. from PEX or the operator)
            // or if the peer tries to connect to us, it may be reintroduced in the future
            val newPeerStateForNode = Peers(node).exclude(Set(a)).mapBy(x => peerState.get(node).get(x))
            peerState' = peerState.set(node, newPeerStateForNode),
            val newConnectionStateForNode = Peers(node).exclude(Set(a)).mapBy(x => connectionState.get(node).get(x))
            connectionState' = connectionState.set(node, newConnectionStateForNode),
        },
        all { //unchanged
            AddressBook' = AddressBook,
            Environment' = Environment,
            IsUnconditional' = IsUnconditional,
            PersistentPeers' = PersistentPeers,
            OperatorAddressBook' = OperatorAddressBook,
        }
    }

    // TODO: this should be a function
    // incoming request
    action transportAccept(node:Address, a: Address):bool = all {
        // If we are connected to the peer as outbound, we do nothing. Otherwise we accept the 
        // connection at this point
        // TODO: Igor thinks about the if below and whether it should/can be an implication
        if ((peersInbound(node) < MaxNumInboundPeers or IsUnconditional.get(node).contains(a)) 
            // under the negation of this condition the connection is rejected
            // TODO: Does rejection unfluence the local state?
            and
            Peers(node).contains(a) and peerState.get(node).get(a) != "outbound") all {
            peerState' = putPeerState(node, a, "inbound"), 
            connectionState' = putConnectionState(node, a, "open"),
            // TODO: inform reactors. Do we model this?
            AddressBook' = AddressBook.setBy(node, (old => old.union(Set(a)))),
            all { //unchanged
                Environment' = Environment,
                IsUnconditional' = IsUnconditional,
                PersistentPeers' = PersistentPeers,
                OperatorAddressBook' = OperatorAddressBook,
            },
            }
            else allUnchanged, 
    }

    // lower level connection initialization complete. Gloss over details
    // TODO: what happens if the other side doesn't accept the connection of some problem occurs?
    action establishConnection(node:Address, a:Address):bool = all { // TODO: upgradeConnection?
        require(or {
            peerState.get(node).get(a) == "dialing",
            peerState.get(node).get(a) == "reconnecting",
            }
        ),
        peerState' = setPeerState(node, a, "outbound"),
        connectionState' = setConnectionState(node, a, "open"),
        all { //unchanged
            AddressBook' = AddressBook,
            Environment' = Environment,
            IsUnconditional' = IsUnconditional,
            PersistentPeers' = PersistentPeers,
            OperatorAddressBook' = OperatorAddressBook,
        }
    }

    // TODO: what to do if reconnection failed?

    action dialingFailed(node:Address, a:Address):bool = all{ 
        require(connectionState.get(node).get(a) == "dialing"),
        val newConnectionStateForNode = Peers(node).exclude(Set(a)).mapBy(x => connectionState.get(node).get(x))
        connectionState' = connectionState.set(node, newConnectionStateForNode),
        val newPeerStateForNode = Peers(node).exclude(Set(a)).mapBy(x => peerState.get(node).get(x))
        peerState' = peerState.set(node, newPeerStateForNode),
        all { //unchanged
            AddressBook' = AddressBook,
            Environment' = Environment,
            IsUnconditional' = IsUnconditional,
            PersistentPeers' = PersistentPeers,
            OperatorAddressBook' = OperatorAddressBook,
        }
   }

    // TODO: this should be a function
    // triggered by a reactor
    action stopPeerForError(node:Address, a:Address):bool = all {
        // TODO: do we care if a is not a peer anymore, e.g. if some other reactor already invkoked
        //       this function? 
        val newPeerStateForNode = Peers(node).exclude(Set(a)).mapBy(x => peerState.get(node).get(x))
        peerState' = peerState.set(node, newPeerStateForNode),
        val newConnectionStateForNode = Peers(node).exclude(Set(a)).mapBy(x => connectionState.get(node).get(x))
        connectionState' = connectionState.set(node, newConnectionStateForNode),
        all { //unchanged
            AddressBook' = AddressBook,
            Environment' = Environment,
            IsUnconditional' = IsUnconditional,
            PersistentPeers' = PersistentPeers,
            OperatorAddressBook' = OperatorAddressBook,
        }
        // TODO: inform all reactors then. (How) should we model that?
    }

       // TODO: this should be a function
    action dialPeer(node:Address, a:Address):bool = all {
        // TOOO: needs to become an if with stuttering if precondition is violated
        require(peersOutbound(node) + peersDialing(node) < MaxNumOutboundPeers),
        require(not(Peers(node).contains(a))), // TODO: English spec restricted to connected or dialing
        all {
            //dialPeer(addr)
            peerState' = putPeerState(node, a, "dialing"), 
            connectionState' = putConnectionState(node, a, "init"),
            all { //unchanged
                AddressBook' = AddressBook,
                Environment' = Environment,
                IsUnconditional' = IsUnconditional,
                PersistentPeers' = PersistentPeers,
                OperatorAddressBook' = OperatorAddressBook,
            }
        }
    }

   // TODO: this should be a function
    action ensurePeers(node:Address):bool = all { 
        /* 
        TODO:
        - shall we pick a subset of Addressbook and dial them
          TODO: ask Igor: can I pick a subset here and use dialPeer? would fold work?
        - shall we model dialattempts, etc?
        */
        // TODO: If. In case precondition is violated we set dialSeeds flag
        require(allAddressesArePeers(node) == false),
        nondet a = oneOf(AddressBook.get(node).exclude(Peers(node))) // addrbook.PickAddress(bias)
        dialPeer(node, a),
    }

    def setPeerState(node: Address, old: (Address -> str) => ): Address -> Address -> str = {
        peerState.setBy(node, ())
    }

   // TODO: this should be a function
    // this can be called by the operator
    action dialPeers(node:Address, S:Set[Address]):bool = all {
        // TODO: check what the code actually is doing. Now we don't override existing peers in this spec
        // TODO: check the following: the intention here is NOT to call "dialPeer" for each address in S
        //       in order to not check against MaxNumOutboundPeers

        peerState' = Peers.union(S).mapBy(a => if (Peers.contains(a)) peerState.get(a) else "dialing"), 
        connectionState' = Peers.union(S).mapBy(a => if (Peers.contains(a)) connectionState.get(a) else "init"),

        AddressBook' = AddressBook.setBy(node, (old => old.union(S))),
        all { //unchanged
                Environment' = Environment,
                IsUnconditional' = IsUnconditional,
                PersistentPeers' = PersistentPeers,
                OperatorAddressBook' = OperatorAddressBook,
        }
    }

    // TODO: this should be a function
    // this is a response to receiving a set of addresses from a seed node
    action PexAddrs(S:Set[Address]):bool = all {
        // we check the allAddressesArePeers it set to true. This means PEX asked seed nodes for addresses.
        // only then we acceppt responses
        dialPeers(S)
        // TODO: currently does not check max outbound peers. There is an issue on that
    }


    /*
    Here we start to describe initial states and runs and tests
    */

    action init = all {
        peerState' = Map(),
        connectionState' = Map(),
        AddressBook' = Set({ID: "igor", networkAddress: "127.0.0.1:10"}, {ID: "josef", networkAddress: "127.0.0.2:10"}),
        Environment' = Set({ID: "lasaro", networkAddress: "127.0.0.7:10"}, {ID: "daniel", networkAddress: "127.0.0.6:10"}),
        IsUnconditional' = Set({ID: "zarko", networkAddress: "127.0.0.42:10"}),
        PersistentPeers' = Set({ID: "bucky", networkAddress: "127.0.0.9:10"}),
        OperatorAddressBook' = Set({ID: "alice", networkAddress: "127.0.0.111:10"}, {ID: "bob", networkAddress: "127.0.0.112:10"}),
    }

    action initWithSameSets =
        val initSet = Set({ID: "igor", networkAddress: "127.0.0.1:10"}, {ID: "josef", networkAddress: "127.0.0.2:10"})
        all {
            peerState' = Map(),
            connectionState' = Map(),
            AddressBook' = initSet,
            Environment' = initSet,
            IsUnconditional' = Set({ID: "zarko", networkAddress: "127.0.0.42:10"}),
            PersistentPeers' = Set({ID: "bucky", networkAddress: "127.0.0.9:10"}),
            OperatorAddressBook' = Set({ID: "alice", networkAddress: "127.0.0.111:10"}, {ID: "bob", networkAddress: "127.0.0.112:10"}),
        }

    action step = any {
        // This is called periodically by PEX
        ensurePeers,

        all { 
            require(Peers != Set()),
            nondet a = oneOf(Peers) 
            any {            
                // dialPeer and dialPeers start the process of connecting to a peer. 
                // dialPeer(a) is no top level action but it is called via via ensurePeers in PEX
                // Depending on the environment this can be succesful (establishconnection) 
                // or not (dialingFailed)
                establishConnection(a),
                dialingFailed(a),

                // at any point of the lifecycle of a peers we may lose the connection due to
                // reasons external to a node (the peer crashes, timeout, disconnection etc.)
                // once we sense that the connection is lost we perform onLostConnection
                loseConnection(a),
                onLostConnection(a),

                // If a reactor is not happy with a peer anymore (e.g., it sent faulty information)
                // the reactor can invoke stopPeerForError and the peer gets removed. 
                stopPeerForError(a),
            }
        },    
        
        // the operator calls dialPeers with her private address book
        nondet S = oneOf(OperatorAddressBook.powerset())
        dialPeers(S),

        // some seed nodes sent us a set S upon request
        nondet S = oneOf(Environment.powerset())
        all {
            // This is a hack. Actually I want to set and re-set a boolean
            require(allAddressesArePeers),
            PexAddrs(S),
        },
        // This is a reaction to incoming requests from the environment
        all {
            require(Environment != Set()),
            nondet a = oneOf(Environment) 
            transportAccept(a),
        }
    }

    

    run happyPathTest = {
        val a = {ID: "alice", networkAddress: "127.0.0.111:10"}
        init
        .then(dialPeers(Set(a)))
        .then(establishConnection(a))
        .then(all{
            assert(connectionState.get(a) == "open" and peerState.get(a) == "outbound"),
            allUnchanged,
        })
    }

    run ensureHappyTest = {
        init
        .then(ensurePeers)
        .then(all{
            assert(Peers != Set()),
            allUnchanged
        })
    }

    run dontAcceptIncomingIfOutboundTest = {
        run pickAAndRun = {
            // nondet a = oneOf(Environment)
            val a = {ID: "igor", networkAddress: "127.0.0.1:10"}
            dialPeer(a)
            .then(establishConnection(a))
            .then(transportAccept(a))
            .then(all{
                assert (peerState.get(a) == "outbound"),
                allUnchanged})}

        initWithSameSets
        .then(pickAAndRun)
    }

}