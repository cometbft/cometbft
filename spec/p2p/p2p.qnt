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
    var peerState: Address -> str
    var connectionState: Address -> str
    var IsUnconditional: Set[Address]

    // PEX variable
    var AddressBook: Set[Address]
    var PersistentPeers: Set[Address] // typically outbound.

    // Environment. These are the nodes that may dial us
    var Environment: Set[Address]
 
    // What we call Peers is defined the domain of the peerState and connectionState maps. 
    // They should be the same, so we can check the invariant invPeers
    def Peers = peerState.keys()
    // auxiliary definition. These are the peers that seem alive and connected
    def activePeers = Peers.filter(x => connectionState.get(x) == "open")

    val invPeers = peerState.keys() == connectionState.keys()
    val invConnectionsNotOpen = Peers.forall(x => 
        (peerState.get(x) == "dialing" or peerState.get(x) == "reconnecting") implies connectionState.get(x) != "open")
    val invConnectionsNotInit = Peers.forall(x => 
        (peerState.get(x) == "inbound" or peerState.get(x) == "outbound") implies connectionState.get(x) != "init")
    val invConnectionsOpen = activePeers.forall(x => peerState.get(x) == "outbound" or peerState.get(x) == "inbound")

    // counting peers in different states
    def peersOutbound = size(Peers.filter(a => peerState.get(a) == "outbound"))
    def peersInbound = size(Peers.filter(a => peerState.get(a) == "inbound"))
    def peersDialing = size(Peers.filter(a => peerState.get(a) == "dialing"))

    // auxiliary definition for PEX, If we don't have peers in the addressbook to try, we dial Seeds to ask for new ones.
    def dialSeeds = AddressBook.exclude(Peers) == Set()

    // helps for better readability. Future syntactic sugar 
    def require(pred:bool):bool = pred
    

    /*
    We will distinguish three kinds of actions according to their origin
    - low level by "the network". This includes incoming requests, connections being lost, but also
      the result of low-level protocol executions (e.g., setting a connection to connected after it was 
      succesful established)
    - reactors: reactors can ask p2p to remove peers if they fail w.r.t. reactor protocol logic
    - PEX reactor: This is a specific reactor that makes sure "enough" connections are established.
    */

    // triggered by "the network"
    action loseConnection(a:Address):bool = all {
        require(connectionState.get(a) != "closed"),
        connectionState' = connectionState.set(a, "closed"),
        all { //unchanged
            AddressBook' = AddressBook,
            peerState' = peerState,
            Environment' = Environment,
            IsUnconditional' = IsUnconditional,
            PersistentPeers' = PersistentPeers,
        }
    }

    // TODO: not sure what triggers that in the code.
    // We try to reconnect to an existing peer, otherwise we just clean up
    action onLostConnection(a:Address):bool = all {
        require(connectionState.get(a) == "closed"),
        if (PersistentPeers.contains(a)) all {
            peerState' = peerState.set(a, "reconnecting"),
            connectionState' = connectionState.set(a, "dialing"),
        }
        else all {
            peerState' = Peers.exclude(Set(a)).mapBy(x => peerState.get(x)),
            connectionState' = Peers.exclude(Set(a)).mapBy(x => connectionState.get(x)),
        },
        all { //unchanged
            AddressBook' = AddressBook,
            Environment' = Environment,
            IsUnconditional' = IsUnconditional,
            PersistentPeers' = PersistentPeers,
        }
    }


    // incoming request
    action transportAccept(a: Address):bool = all {
        or { // under the negation of this condition the connection is rejected
             // TODO: Does rejection unfluence the local state?
            peersInbound < MaxNumInboundPeers,
            IsUnconditional.contains(a)
        },
        
        require(if (Peers.contains(a)) peerState.get(a) != "outbound" 
                else true), // TODO: check whether is is the right condition
        peerState' = peerState.put(a, "inbound"), 
        connectionState' = connectionState.put(a, "open"),
        // TODO: inform reactors. Do we model this?
        AddressBook' = AddressBook.union(Set(a)) ,
        all { //unchanged
            Environment' = Environment,
            IsUnconditional' = IsUnconditional,
            PersistentPeers' = PersistentPeers,
        }
    }

    // lower level connection initialization complete. Gloss over details
    // TODO: what happens if the other side doesn't accept the connection of some problem occurs?
    action establishConnection(a:Address):bool = all { // TODO: upgradeConnection?
        require(or {
            peerState.get(a) == "dialing",
            peerState.get(a) == "reconnecting",
            }
        ),
        peerState' = peerState.set(a, "outbound"),
        connectionState' = connectionState.set(a, "open"),
        all { //unchanged
            AddressBook' = AddressBook,
            Environment' = Environment,
            IsUnconditional' = IsUnconditional,
            PersistentPeers' = PersistentPeers,
        }
    }

    action dialingFailed(a:Address):bool = all{ 
        require(connectionState.get(a) == "dialing"),
        connectionState' = Peers.exclude(Set(a)).mapBy(x => connectionState.get(x)),
        peerState' = Peers.exclude(Set(a)).mapBy(x => peerState.get(x)),
        all { //unchanged
            AddressBook' = AddressBook,
            Environment' = Environment,
            IsUnconditional' = IsUnconditional,
            PersistentPeers' = PersistentPeers,
        }
   }

    // triggered by a reactor
    action stopPeerForError(a:Address):bool = all {
        peerState' = Peers.exclude(Set(a)).mapBy(x => peerState.get(x)),
        connectionState' = Peers.exclude(Set(a)).mapBy(x => connectionState.get(x)),
        all { //unchanged
            AddressBook' = AddressBook,
            Environment' = Environment,
            IsUnconditional' = IsUnconditional,
            PersistentPeers' = PersistentPeers,
        }
        // TODO: inform all reactors then. (How) should we model that?
    }

    // triggered by PEX
    action debugEnsurePeers(a:Address):bool = all {
    // this is for debugging the spec. In the protocol EnsurePeers does a random pick of an address
    // here we already entered with a specific address.
        require(peersOutbound + peersDialing < MaxNumOutboundPeers),
        require(not(Peers.contains(a))), // TODO: English spec restricted to connected or dialing
        all {
            //dialPeer(addr)
            peerState' = peerState.put(a, "dialing"), 
            connectionState' = connectionState.put(a, "init"),
            all { //unchanged
                AddressBook' = AddressBook,
                Environment' = Environment,
                IsUnconditional' = IsUnconditional,
                PersistentPeers' = PersistentPeers,
            }
        }
    }


    action ensurePeers = all { 
        /* 
        TODO:
        - shall we pick a subset of Addressbook and dial them
        - shall we model dialattempts, etc?
        */
        require(dialSeeds == false),
        nondet a = oneOf(AddressBook.exclude(Peers)) // addrbook.PickAddress(bias)
        debugEnsurePeers(a),
    }

    // this can be called by the operator
    action dialPeers(S:Set[Address]):bool = all {
        // TODO: check what the code actually is doing. Now we don't override existing peers in this spec
        peerState' = Peers.union(S).mapBy(a => if (Peers.contains(a)) peerState.get(a) else "dialing"), 
        connectionState' = Peers.union(S).mapBy(a => if (Peers.contains(a)) connectionState.get(a) else "init"),
        AddressBook' = AddressBook.union(S),
        all { //unchanged
                Environment' = Environment,
                IsUnconditional' = IsUnconditional,
                PersistentPeers' = PersistentPeers,
        }
    }

    // this is a response to receiving a set of addresses from a seed node
    action PexAddrs(S:Set[Address]):bool = all {
        require(dialSeeds),
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
        }

    action step = any {
        ensurePeers,
        all {
            require(Peers != Set()),
            nondet a = oneOf(Peers) 
            stopPeerForError(a),
        },
        all {
            require(Peers != Set()),
            nondet a = oneOf(Peers) 
            loseConnection(a),
        },
        all {
            require(Peers != Set()),
            nondet a = oneOf(Peers) 
            dialingFailed(a),
        }, 
        all {
            require(Peers != Set()),
            nondet a = oneOf(Peers) 
            establishConnection(a),
        },    
        all {
            require(Environment != Set()),
            nondet a = oneOf(Environment) 
            transportAccept(a),
        }
    }

    


}