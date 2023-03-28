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
    pure val peerStates = Set("reconnecting", "inbound", "outbound")
    pure val connectionStates = Set("incoming", "outgoing")
    // TODO: We might experiment around the max peer numbers. Should we model them as variables?
    pure val MaxNumOutboundPeers = 5
    pure val MaxNumInboundPeers = 10

    // (public_key, ipAddress:port)
    type Address = {ID:str, networkAddress:str}

    // Environment. These are the nodes that may dial us
    // TODO: should be constant (set via instances). Doesn't work, yet.
    var Environment: Set[Address]
 
    // Operator. The operator can manually provide a list of addresses to dial via dialPeers
    // TODO: should be constant (set via instances). Doesn't work, yet.
    var OperatorAddressBook: Set[Address]

    type Connection = {
        address : Address,
        direction: str, // "incoming" "outgoing"
    }

    type LocalState = {
        // p2p variables 
        peerState: Address -> str,
        connections: Set[Connection],
        IsUnconditional: Set[Address],
        // PEX variable
        AddressBook: Set[Address],
        PersistentPeers: Set[Address],
    }

    var config: Address -> LocalState

    pure def Peers(s: LocalState): Set[Address] = s.peerState.keys()

    def modeledNodes = config.keys()

    // TODO: write functions as definitions.
    // Then write actions that use the definitions, and map "local" defintions as they act on the global statte
    // Write actions that describe the environment

    // helps for better readability. Future syntactic sugar 
    def require(pred:bool):bool = pred


    pure def isDialing(s: LocalState, a:Address) : bool = and {
        s.connections.contains({address: a, direction: "outgoing"}),
        or {
            not(s.Peers().contains(a)),
            and {
                // we are dialing to 'a', but in the meantime we accepted 'a' from an incoming connection
                s.Peers().contains(a),
                s.peerState.get(a) == "inbound",
            }
        }
    }
    // counting peers (actually addresses) in different states
    pure def peersOutbound(s: LocalState): int = size(s.Peers().filter(a => s.peerState.get(a) == "outbound"))
    pure def peersInbound(s: LocalState): int = size(s.Peers().filter(a => s.peerState.get(a) == "inbound"))
    pure def peersDialing(s: LocalState): int = size(s.connections.filter(c => s.isDialing(c.address)))



// TODO: ecode these workflow
 // callDialPeer(Address) ->  makes an outgoing connection 

 // return DialPeerFailed -> delete connection from map

/*
 * p2p peer management functions
 */

    // This calls a potential peer. This means some lower level network code is started, and communication
    // with the potential peer is started. Eventually this will result in one of the actions
    // lowerLevelDialPeerSuccess or lowerLevelDialPeerFailed. See below.
    pure def dialPeer(s: LocalState, a: Address): LocalState = {
        if (s.peersOutbound() + s.peersDialing() < MaxNumOutboundPeers and not(s.Peers().contains(a))) {
            val newConn : Set[Connection] = s.connections.union(Set({address: a, direction: "outgoing"}))
            s.with("connections", newConn)
        }
        else {
            s
        }
    }    

 // transportAccept(Address) ->  makes an incoming connection 
     pure def transportAccept(s: LocalState, a: Address): LocalState = {
        if (s.peersInbound() < MaxNumInboundPeers or s.IsUnconditional.contains(a)) {
            val newConn : Set[Connection] = s.connections.union(Set({address: a, direction: "incoming"}))
            val newAB : Set[Address] = s.AddressBook.union(Set(a))
            s.with("connections", newConn)
             .with("AddressBook", newAB)
        }
        else {
            s
        }
    }    

/*
 * auxiliary actions. Perhaps local
 */

   // node dials a. This starts lower lever computations. 
    action actionDialPeer (node: Address, a:Address):bool = all {
        config' = config.set(node, config.get(node).dialPeer(a)),
        all { //unchanged
            Environment' = Environment,
            OperatorAddressBook' = OperatorAddressBook,
        }
    }

    action actionTransportAccept (node: Address, a:Address):bool = all {
        config' = config.set(node, config.get(node).transportAccept(a)),
        all { //unchanged
            Environment' = Environment,
            OperatorAddressBook' = OperatorAddressBook,
        }
    }


/* 
 * Environment actions
 */

    // auxiliary functions
    pure def removeConnectionOutgoing(s: LocalState, a: Address): LocalState = {
        val y : Set[Connection] = s.connections.exclude(Set({address: a, direction: "outgoing"}))
        s.with("connections", y)
    }

    pure def putPeer(s: LocalState, a: Address, direction: str): LocalState = {
        val y =  s.peerState.put(a, direction)
        s.with("peerState", y)
    }

 // return DialPeerSuccess -> if connection.address.id not in Peers -> peerState(address) = "outbound" 
 //                           otherwise -> delete connection from map 
    action lowerLevelDialPeerSuccess(node: Address, a: Address): bool = all {
        require(config.get(node).connections.contains({address: a, direction: "outgoing"})), 
        if (config.get(node).Peers().filter(x => x.ID == a.ID) == Set()) all {
            // If I don't have the ID in my peers already, add the address
            config' = config.set(node, config.get(node).putPeer(a, "outbound"))
        }
        else {
            // I have the ID already. That is, it might be incoming, or I might have dialed over a different
            // (network) address
            config' = config.set(node, config.get(node).removeConnectionOutgoing(a))
        },
        all { //unchanged
            Environment' = Environment,
            OperatorAddressBook' = OperatorAddressBook,
        }
    }

    action lowerLevelDialPeerFailed(node: Address, a: Address): bool = all {
        require(config.get(node).connections.contains({address: a, direction: "outgoing"})), 
        // TODO: perhaps add some checks
        config' = config.set(node, config.get(node).removeConnectionOutgoing(a)),
        all { //unchanged
            Environment' = Environment,
            OperatorAddressBook' = OperatorAddressBook,
        }
    }

   action lowerLevelTransportAcceptSuccess(node: Address, a: Address): bool = all {
        require(config.get(node).connections.contains({address: a, direction: "incoming"})), 
        if (config.get(node).Peers().filter(x => x.ID == a.ID) == Set()) all {
            // If I don't have the ID in my peers already, add the address
            config' = config.set(node, config.get(node).putPeer(a, "inbound"))
        }
        else {
            // I have the ID already. That is, it might be incoming, or I might have dialed over a different
            // (network) address
            config' = config.set(node, config.get(node).removeConnectionOutgoing(a))
        },
        all { //unchanged
            Environment' = Environment,
            OperatorAddressBook' = OperatorAddressBook,
        }
    }

    action lowerLevelTransportAcceptFailed(node: Address, a: Address): bool = all {
        require(config.get(node).connections.contains({address: a, direction: "incoming"})), 
        // TODO: perhaps add some checks
        config' = config.set(node, config.get(node).removeConnectionOutgoing(a)),
        all { //unchanged
            Environment' = Environment,
            OperatorAddressBook' = OperatorAddressBook,
        }
    }





/*
 * THE REST
 */


    // val invPeers = modeledNodes.forall(y => config.get(y).peerState.keys() == config.get(y).connectionState.keys())

    // // auxiliary definition. These are the peers that seem alive and connected
    // def activePeers (s: LocalState): Set[Address] = s.Peers().filter(x => s.connectionState.get(x) == "open")

    // // for all modeled nodes y and all its peers x, some peerStates imply some connection state
    // val invConnectionsNotOpen = modeledNodes.forall(y => config.get(y).Peers().forall(x => 
    //     (config.get(y).peerState.get(x) == "dialing" or config.get(y).peerState.get(x) == "reconnecting") implies 
    //         config.get(y).connectionState.get(x) != "open"))

    // val invConnectionsNotInit = modeledNodes.forall(y => config.get(y).Peers().forall(x => 
    //     (config.get(y).peerState.get(x) == "inbound" or config.get(y).peerState.get(x) == "outbound") implies 
    //         config.get(y).connectionState.get(x) != "init"))

    // val invConnectionsOpen = modeledNodes.forall(y => config.get(y).activePeers().forall(x => 
    //     config.get(y).peerState.get(x) == "outbound" or config.get(y).peerState.get(x) == "inbound"))



//     // auxiliary definition for PEX, 
//     // If we don't have peers in the addressbook to try, we dial Seeds to ask for new ones.
//     def allAddressesArePeers(p) = AddressBook.get(p).exclude(Peers(p)) == Set()


    // This is painful
    action allUnchanged = all {
        config' = config,
        Environment' = Environment,
        OperatorAddressBook' = OperatorAddressBook,
    }    



 

//     /*
//     We will distinguish three kinds of actions according to their origin
//     - low level by "the network". This includes incoming requests, connections being lost, but also
//       the result of low-level protocol executions (e.g., setting a connection to connected after it was 
//       succesful established)
//     - reactors: reactors can ask p2p to remove peers if they fail w.r.t. reactor protocol logic
//     - PEX reactor: This is a specific reactor that makes sure "enough" connections are established.
//     */

    

    //  def setConnectionState(s: LocalState, a: Address, newval: str) : LocalState = {
    //     s.with("connectionState", s.connectionState.set(a, newval))
    //  }

//     def setPeerState(node, a, newval) = 
//         peerState.setBy(node, (old => old.set(a, newval)))
 
//     def putConnectionState(node, a, newval) = 
//         connectionState.setBy(node, (old => old.put(a, newval)))

//     def putPeerState(node, a, newval) = 
//         peerState.setBy(node, (old => old.put(a, newval)))
    

    // triggered by "the network"
    // action loseConnection(node: Address, a:Address):bool = all {
    //     require(config.get(node).connectionState.get(a) != "closed"),
    //     // The following line is obscure magic and we discovered a bug in the cheat sheet
    //     config' = config.set(node, config.get(node).setConnectionState(a, "closed") ),
    //     all { //unchanged
    //         Environment' = Environment,
    //         OperatorAddressBook' = OperatorAddressBook,
    //     }
    // }

//     // TODO: not sure what triggers that in the code.
//     // We try to reconnect to an existing peer, otherwise we just clean up
//     action onLostConnection(node: Address, a:Address):bool = all {
//         require(connectionState.get(node).get(a) == "closed"),
//         if (PersistentPeers.get(node).contains(a)) all {
//             peerState' = setPeerState(node, a, "reconnecting"),
//             connectionState' = setConnectionState(node, a, "dialing"),
//         }
//         else all {
//             // if peer is not persistent, we just delete it
//             // it may still be in some addressbook. Thus upon request (e.g. from PEX or the operator)
//             // or if the peer tries to connect to us, it may be reintroduced in the future
//             val newPeerStateForNode = Peers(node).exclude(Set(a)).mapBy(x => peerState.get(node).get(x))
//             peerState' = peerState.set(node, newPeerStateForNode),
//             val newConnectionStateForNode = Peers(node).exclude(Set(a)).mapBy(x => connectionState.get(node).get(x))
//             connectionState' = connectionState.set(node, newConnectionStateForNode),
//         },
//         all { //unchanged
//             AddressBook' = AddressBook,
//             Environment' = Environment,
//             IsUnconditional' = IsUnconditional,
//             PersistentPeers' = PersistentPeers,
//             OperatorAddressBook' = OperatorAddressBook,
//         }
//     }

//     // TODO: this should be a function
//     // incoming request
//     action transportAccept(node:Address, a: Address):bool = all {
//         // If we are connected to the peer as outbound, we do nothing. Otherwise we accept the 
//         // connection at this point
//         // TODO: Igor thinks about the if below and whether it should/can be an implication
//         if ((peersInbound(node) < MaxNumInboundPeers or IsUnconditional.get(node).contains(a)) 
//             // under the negation of this condition the connection is rejected
//             // TODO: Does rejection unfluence the local state?
//             and
//             Peers(node).contains(a) and peerState.get(node).get(a) != "outbound") all {
//             peerState' = putPeerState(node, a, "inbound"), 
//             connectionState' = putConnectionState(node, a, "open"),
//             // TODO: inform reactors. Do we model this?
//             AddressBook' = AddressBook.setBy(node, (old => old.union(Set(a)))),
//             all { //unchanged
//                 Environment' = Environment,
//                 IsUnconditional' = IsUnconditional,
//                 PersistentPeers' = PersistentPeers,
//                 OperatorAddressBook' = OperatorAddressBook,
//             },
//             }
//             else allUnchanged, 
//     }

//     // lower level connection initialization complete. Gloss over details
//     // TODO: what happens if the other side doesn't accept the connection of some problem occurs?
//     action establishConnection(node:Address, a:Address):bool = all { // TODO: upgradeConnection?
//         require(or {
//             peerState.get(node).get(a) == "dialing",
//             peerState.get(node).get(a) == "reconnecting",
//             }
//         ),
//         peerState' = setPeerState(node, a, "outbound"),
//         connectionState' = setConnectionState(node, a, "open"),
//         all { //unchanged
//             AddressBook' = AddressBook,
//             Environment' = Environment,
//             IsUnconditional' = IsUnconditional,
//             PersistentPeers' = PersistentPeers,
//             OperatorAddressBook' = OperatorAddressBook,
//         }
//     }

//     // TODO: what to do if reconnection failed?

//     action dialingFailed(node:Address, a:Address):bool = all{ 
//         require(connectionState.get(node).get(a) == "dialing"),
//         val newConnectionStateForNode = Peers(node).exclude(Set(a)).mapBy(x => connectionState.get(node).get(x))
//         connectionState' = connectionState.set(node, newConnectionStateForNode),
//         val newPeerStateForNode = Peers(node).exclude(Set(a)).mapBy(x => peerState.get(node).get(x))
//         peerState' = peerState.set(node, newPeerStateForNode),
//         all { //unchanged
//             AddressBook' = AddressBook,
//             Environment' = Environment,
//             IsUnconditional' = IsUnconditional,
//             PersistentPeers' = PersistentPeers,
//             OperatorAddressBook' = OperatorAddressBook,
//         }
//    }

//     // TODO: this should be a function
//     // triggered by a reactor
//     action stopPeerForError(node:Address, a:Address):bool = all {
//         // TODO: do we care if a is not a peer anymore, e.g. if some other reactor already invkoked
//         //       this function? 
//         val newPeerStateForNode = Peers(node).exclude(Set(a)).mapBy(x => peerState.get(node).get(x))
//         peerState' = peerState.set(node, newPeerStateForNode),
//         val newConnectionStateForNode = Peers(node).exclude(Set(a)).mapBy(x => connectionState.get(node).get(x))
//         connectionState' = connectionState.set(node, newConnectionStateForNode),
//         all { //unchanged
//             AddressBook' = AddressBook,
//             Environment' = Environment,
//             IsUnconditional' = IsUnconditional,
//             PersistentPeers' = PersistentPeers,
//             OperatorAddressBook' = OperatorAddressBook,
//         }
//         // TODO: inform all reactors then. (How) should we model that?
//     }

//        // TODO: this should be a function
//     action dialPeer(node:Address, a:Address):bool = all {
//         // TOOO: needs to become an if with stuttering if precondition is violated
//         require(peersOutbound(node) + peersDialing(node) < MaxNumOutboundPeers),
//         require(not(Peers(node).contains(a))), // TODO: English spec restricted to connected or dialing
//         all {
//             //dialPeer(addr)
//             peerState' = putPeerState(node, a, "dialing"), 
//             connectionState' = putConnectionState(node, a, "init"),
//             all { //unchanged
//                 AddressBook' = AddressBook,
//                 Environment' = Environment,
//                 IsUnconditional' = IsUnconditional,
//                 PersistentPeers' = PersistentPeers,
//                 OperatorAddressBook' = OperatorAddressBook,
//             }
//         }
//     }

//    // TODO: this should be a function
//     action ensurePeers(node:Address):bool = all { 
//         /* 
//         TODO:
//         - shall we pick a subset of Addressbook and dial them
//           TODO: ask Igor: can I pick a subset here and use dialPeer? would fold work?
//         - shall we model dialattempts, etc?
//         */
//         // TODO: If. In case precondition is violated we set dialSeeds flag
//         require(allAddressesArePeers(node) == false),
//         nondet a = oneOf(AddressBook.get(node).exclude(Peers(node))) // addrbook.PickAddress(bias)
//         dialPeer(node, a),
//     }

//     def setPeerState(node: Address, old: (Address -> str) => ): Address -> Address -> str = {
//         peerState.setBy(node, ())
//     }

//    // TODO: this should be a function
//     // this can be called by the operator
//     action dialPeers(node:Address, S:Set[Address]):bool = all {
//         // TODO: check what the code actually is doing. Now we don't override existing peers in this spec
//         // TODO: check the following: the intention here is NOT to call "dialPeer" for each address in S
//         //       in order to not check against MaxNumOutboundPeers

//         peerState' = Peers.union(S).mapBy(a => if (Peers.contains(a)) peerState.get(a) else "dialing"), 
//         connectionState' = Peers.union(S).mapBy(a => if (Peers.contains(a)) connectionState.get(a) else "init"),

//         AddressBook' = AddressBook.setBy(node, (old => old.union(S))),
//         all { //unchanged
//                 Environment' = Environment,
//                 IsUnconditional' = IsUnconditional,
//                 PersistentPeers' = PersistentPeers,
//                 OperatorAddressBook' = OperatorAddressBook,
//         }
//     }

//     // TODO: this should be a function
//     // this is a response to receiving a set of addresses from a seed node
//     action PexAddrs(S:Set[Address]):bool = all {
//         // we check the allAddressesArePeers it set to true. This means PEX asked seed nodes for addresses.
//         // only then we acceppt responses
//         dialPeers(S)
//         // TODO: currently does not check max outbound peers. There is an issue on that
//     }


    /*
    Here we start to describe initial states and runs and tests
    */

    action init = all {
        OperatorAddressBook' = Set({ID: "alice", networkAddress: "127.0.0.111:10"}, {ID: "bob", networkAddress: "127.0.0.112:10"}),
        Environment' = Set({ID: "lasaro", networkAddress: "127.0.0.7:10"}, {ID: "daniel", networkAddress: "127.0.0.6:10"}),
        config' = Map (
            {ID: "scully", networkAddress: "127.0.0.1177:10"} -> {
                peerState: Map(),
                connections: Set(),
                AddressBook: Set({ID: "igor", networkAddress: "127.0.0.1:10"}, {ID: "josef", networkAddress: "127.0.0.2:10"}),
                IsUnconditional: Set({ID: "zarko", networkAddress: "127.0.0.42:10"}),
                PersistentPeers: Set({ID: "bucky", networkAddress: "127.0.0.9:10"}),
            }
        )
    }

    action step = any {
        nondet a = oneOf(modeledNodes) 
        nondet outgoing = oneOf(config.get(a).AddressBook)
        nondet incoming = oneOf(Environment)
        any {            
            actionDialPeer(a, outgoing),
            lowerLevelDialPeerFailed(a, outgoing),
            lowerLevelDialPeerSuccess(a, outgoing),
            actionTransportAccept(a, incoming),
            lowerLevelTransportAcceptFailed(a, incoming),
            lowerLevelDialPeerSuccess(a, incoming)
        }
    }

    // action initWithSameSets =
    //     val initSet = Set({ID: "igor", networkAddress: "127.0.0.1:10"}, {ID: "josef", networkAddress: "127.0.0.2:10"})
    //     all {
    //         peerState' = Map(),
    //         connectionState' = Map(),
    //         AddressBook' = initSet,
    //         Environment' = initSet,
    //         IsUnconditional' = Set({ID: "zarko", networkAddress: "127.0.0.42:10"}),
    //         PersistentPeers' = Set({ID: "bucky", networkAddress: "127.0.0.9:10"}),
    //         OperatorAddressBook' = Set({ID: "alice", networkAddress: "127.0.0.111:10"}, {ID: "bob", networkAddress: "127.0.0.112:10"}),
    //     }
 
    // action step = any {
    //     nondet a = oneOf(modeledNodes) 
    //     nondet d = oneOf(config.get(a).AddressBook)
    //     actionDialPeer(a,d)
    // }



    

    // run happyPathTest = {
    //     val a = {ID: "alice", networkAddress: "127.0.0.111:10"}
    //     init
    //     .then(dialPeers(Set(a)))
    //     .then(establishConnection(a))
    //     .then(all{
    //         assert(connectionState.get(a) == "open" and peerState.get(a) == "outbound"),
    //         allUnchanged,
    //     })
    // }

    // run ensureHappyTest = {
    //     init
    //     .then(ensurePeers)
    //     .then(all{
    //         assert(Peers != Set()),
    //         allUnchanged
    //     })
    // }

    // run dontAcceptIncomingIfOutboundTest = {
    //     run pickAAndRun = {
    //         // nondet a = oneOf(Environment)
    //         val a = {ID: "igor", networkAddress: "127.0.0.1:10"}
    //         dialPeer(a)
    //         .then(establishConnection(a))
    //         .then(transportAccept(a))
    //         .then(all{
    //             assert (peerState.get(a) == "outbound"),
    //             allUnchanged})}

    //     initWithSameSets
    //     .then(pickAAndRun)
    // }

}