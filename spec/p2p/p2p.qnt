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
    var peerState: Address -> str
    var connectionState: Address -> str
    var IsUnconditional: Set[Address]

    // PEX variable
    var AddressBook: Set[Address]
    var PersistentPeers: Set[Address] // typically outbound.

    // Environment. These are the nodes that may dial us
    var Environment: Set[Address]
 
    // Operator. The operator can manually provide a list of addresses to dial via dialPeers
    var OperatorAddressBook: Set[Address]

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

    // auxiliary definition for PEX, 
    // If we don't have peers in the addressbook to try, we dial Seeds to ask for new ones.
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
            OperatorAddressBook' = OperatorAddressBook,
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
            // if peer is not persistent, we just delete it
            // it may still be in some addressbook. Thus upon request (e.g. from PEX or the operator)
            // or if the peer tries to connect to us, it may be reintroduced in the future
            peerState' = Peers.exclude(Set(a)).mapBy(x => peerState.get(x)),
            connectionState' = Peers.exclude(Set(a)).mapBy(x => connectionState.get(x)),
        },
        all { //unchanged
            AddressBook' = AddressBook,
            Environment' = Environment,
            IsUnconditional' = IsUnconditional,
            PersistentPeers' = PersistentPeers,
            OperatorAddressBook' = OperatorAddressBook,
        }
    }


    // incoming request
    action transportAccept(a: Address):bool = all {
        or { // under the negation of this condition the connection is rejected
             // TODO: Does rejection unfluence the local state?
            peersInbound < MaxNumInboundPeers,
            IsUnconditional.contains(a)
        },

        // If we are connected to the peer as outbound, we do nothing. Otherwise we accept the 
        // connection at this point
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
            OperatorAddressBook' = OperatorAddressBook,
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
            OperatorAddressBook' = OperatorAddressBook,
        }
    }

    // TODO: what to do if reconnection failed?

    action dialingFailed(a:Address):bool = all{ 
        require(connectionState.get(a) == "dialing"),
        connectionState' = Peers.exclude(Set(a)).mapBy(x => connectionState.get(x)),
        peerState' = Peers.exclude(Set(a)).mapBy(x => peerState.get(x)),
        all { //unchanged
            AddressBook' = AddressBook,
            Environment' = Environment,
            IsUnconditional' = IsUnconditional,
            PersistentPeers' = PersistentPeers,
            OperatorAddressBook' = OperatorAddressBook,
        }
   }

    // triggered by a reactor
    action stopPeerForError(a:Address):bool = all {
        // TODO: do we care if a is not a peer anymore, e.g. if some other reactor already invkoked
        //       this function? 
        peerState' = Peers.exclude(Set(a)).mapBy(x => peerState.get(x)),
        connectionState' = Peers.exclude(Set(a)).mapBy(x => connectionState.get(x)),
        all { //unchanged
            AddressBook' = AddressBook,
            Environment' = Environment,
            IsUnconditional' = IsUnconditional,
            PersistentPeers' = PersistentPeers,
            OperatorAddressBook' = OperatorAddressBook,
        }
        // TODO: inform all reactors then. (How) should we model that?
    }

    action dialPeer(a:Address):bool = all {
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
                OperatorAddressBook' = OperatorAddressBook,
            }
        }
    }


    action ensurePeers = all { 
        /* 
        TODO:
        - shall we pick a subset of Addressbook and dial them
          TODO: ask Igor: can I pick a subset here and use dialPeer? would fold work?
        - shall we model dialattempts, etc?
        */
        require(dialSeeds == false),
        nondet a = oneOf(AddressBook.exclude(Peers)) // addrbook.PickAddress(bias)
        dialPeer(a),
    }

    // this can be called by the operator
    action dialPeers(S:Set[Address]):bool = all {
        // TODO: check what the code actually is doing. Now we don't override existing peers in this spec
        // TODO: check the following: the intention here is NOT to call "dialPeer" for each address in S
        //       in order to not check against MaxNumOutboundPeers
        peerState' = Peers.union(S).mapBy(a => if (Peers.contains(a)) peerState.get(a) else "dialing"), 
        connectionState' = Peers.union(S).mapBy(a => if (Peers.contains(a)) connectionState.get(a) else "init"),
        AddressBook' = AddressBook.union(S),
        all { //unchanged
                Environment' = Environment,
                IsUnconditional' = IsUnconditional,
                PersistentPeers' = PersistentPeers,
                OperatorAddressBook' = OperatorAddressBook,
        }
    }

    // this is a response to receiving a set of addresses from a seed node
    action PexAddrs(S:Set[Address]):bool = all {
        // we check the dialSeeds it set to true. This means PEX asked seed nodes for addresses.
        // only then we acceppt responses
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
        PexAddrs(S),

        // This is a reaction to incoming requests from the environment
        all {
            require(Environment != Set()),
            nondet a = oneOf(Environment) 
            transportAccept(a),
        }
    }



}