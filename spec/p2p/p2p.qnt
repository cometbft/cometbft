/* ****************************************************************************
  This encodes the execution flows of the p2p system of CometBFT for 
  maintaining connections to other
  nodes in the network. It is based on the descriptions in 
  https://github.com/cometbft/cometbft/blob/91f8dc3ea0490fd824d1d3d346fb0da61de2f848/spec/p2p/architecture/peers-connect.md?plain=1#L1

  The funcitionality described here is reactive. It is triggered by
  - The PEX reactor (actually a part of the p2p system)
  - CometBFT reactors
  - connection requests from other nodes in the network
  - node operators
  - lower level code (TCP connection management)

  Josef Widder, Informal Systems, 2023
**************************************************************************** */



module p2p {

/* ****************************************************************************
 *
 * Code spec. This is what is supposed to replace the pseudo code in English
 * specifications.
 *
 * ************************************************************************* */


/* ****************************************************************************
 *
 * Data Types
 *
 * ************************************************************************* */

    // TODO: "pure val" is a work around as we don't have enums yet
    pure val peerStates = Set("reconnecting", "inbound", "outbound")
    pure val connectionStates = Set("incoming", "outgoing", "lost")

    // (public_key, ipAddress:port)
    type Address = {ID:str, networkAddress:str}

    type Connection = {
        address : Address,
        state: str, // "incoming" "outgoing" "lost"
    }

    type LocalState = {
        // p2p configuration
        // TODO: is it config?
        MaxNumOutboundPeers: int,
        MaxNumInboundPeers: int,

        // p2p variables 
        peerState: Address -> str,
        connections: Set[Connection],
        IsUnconditional: Set[Address],

        // PEX variable
        AddressBook: Set[Address],
        PersistentPeers: Set[Address],
    }

/* ****************************************************************************
 *
 * The functions that constitute the peer management. 
 *
 * ************************************************************************* */

    /*
     * p2p peer management: auxiliary functions
     */

    pure def Peers(s: LocalState): Set[Address] = s.peerState.keys()

    pure def activePeers(s: LocalState): Set[Address] = 
        s.Peers().filter(p =>
            and {
                s.peerState.get(p) != "reconnecting",
                or {
                    s.connections.contains({address: p, state: "outgoing"}),
                    s.connections.contains({address: p, state: "incoming"})
                }
            }
           
        )

    pure def isDialing(s: LocalState, a:Address) : bool = and {
        s.connections.contains({address: a, state: "outgoing"}),
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


    /*
    * p2p peer management functions
    */

    // This calls a potential peer. This means some lower level network code is started, and communication
    // with the potential peer is started. Eventually this will result in one of the actions
    // lowerLevelDialPeerSuccess or lowerLevelDialPeerFailed. See below.
    pure def dialPeer(s: LocalState, a: Address): LocalState = {
        if (s.peersOutbound() + s.peersDialing() < s.MaxNumOutboundPeers and not(s.Peers().contains(a))) {
            // TODO: shall we check "not(s.isDialing(a))"?
            val newConn : Set[Connection] = s.connections.union(Set({address: a, state: "outgoing"}))
            s.with("connections", newConn)
        }
        else {
            s
        }
    }    

    // transportAccept(Address) ->  makes an incoming connection. Dual to "dialPeer" but incoming
    pure def transportAccept(s: LocalState, a: Address): LocalState = {
        if (s.peersInbound() < s.MaxNumInboundPeers or s.IsUnconditional.contains(a)) {
            val newConn : Set[Connection] = s.connections.union(Set({address: a, state: "incoming"}))
            val newAB : Set[Address] = s.AddressBook.union(Set(a))
            s.with("connections", newConn)
             .with("AddressBook", newAB)
        }
        else {
            s
        }
    }    



    // a connection was lost. We need to figure out whether to reconnect.
    pure def onLostConnection(s: LocalState, c: Connection): LocalState = {
        if (s.PersistentPeers.contains(c.address) and c.state == "lost") {
            val newPS = s.peerState.set(c.address, "reconnecting")
            // set connection state of c to outgoing (connections is a set)
            val newConn = s.connections.exclude(Set(c)).union(Set({address: c.address, state: "outgoing"}))
            s.with("peerState", newPS)
             .with("connections", newConn)
        }
        else {
            // if peer is not persistent, we just delete it
            // it may still be in some addressbook. Thus upon request (e.g. from PEX or the operator)
            // or if the peer tries to connect to us, it may be reintroduced in the future
            val newPS = s.Peers().exclude(Set(c.address)).mapBy(x => s.peerState.get(x))
            s.with("connections", s.connections.exclude(Set(c)))
             .with("peerState", newPS)
        }
    }

    // TODO: check whether the parameter actually is peerID
    // TODO: check functionality. Here I can give an arbitrary peerID. It will not change the state
    //       if this peerID doesn't exist
    pure def stopPeerForError(s: LocalState, peerID: str): LocalState = {
        val newPS = s.Peers().filter(x => x.ID != peerID).mapBy(x => s.peerState.get(x))
        val newConn = s.connections.filter(x => x.address.ID != peerID)
        s.with("peerState", newPS)
         .with("connections", newConn)
    }

    pure def dialPeers(s: LocalState, addresses: Set[Address]): LocalState = {
        // TODO: check what the code actually is doing. Now we don't override existing peers in this spec
        // TODO: check the following: the intention here is NOT to call "dialPeer" for each address in S
        //       in order to not check against MaxNumOutboundPeers
        val newAdr = addresses.filter(a => not(s.connections.exists(c => c.address == a)))
        // TODO: perhaps we should check for IDs rather than addresses above        
        val newConn = s.connections.union(newAdr.map(a => {address: a, state: "outgoing"}))
        s.with("AddressBook", s.AddressBook.union(addresses))
         .with("connections", newConn)
    }

    pure def PexAddrs(s: LocalState, addresses: Set[Address]): LocalState = {
        s.dialPeers(addresses)
    }



// /* ****************************************************************************
//  * Transition system.
//  * This defines how the above functions can be executed
//  * ************************************************************************* */





/* ****************************************************************************
 * Global state
 * ************************************************************************* */

    // These models the nodes that do perform steps in this specification.
    var config: Address -> LocalState
    def modeledNodes = config.keys()
  
    // Environment. These are the addresses of nodes that may dial us (these nodes only appear as data 
    // in this spec. They don't perform actions)
    // TODO: should be constant (set via instances). Doesn't work, yet.
    var Environment: Set[Address]
 
    // Operator's address book. The operator can manually provide a list of addresses to dial via dialPeers
    // TODO: should be constant (set via instances). Doesn't work, yet.
    var OperatorAddressBook: Set[Address]

    // Those modeled nodes that ran out of unused addresses, and asked for more addresses from seed nodes
    // This is used to model actions by PEX, which we don't capture here explicitely
    var requestingNodes: Set[Address]



/* ****************************************************************************
 * Invariants
 * ************************************************************************* */

    // for each peer there exists a connection    
    val invPeers = modeledNodes.forall(n => 
                   config.get(n).Peers().forall(p => 
                   config.get(n).connections.exists(c => c.address == p)))

    // for each connection there is a peer. Only eventually we will add a peer or remove the connection, 
    // so this is expected to break 
    val wrongConn = modeledNodes.forall(n => 
                    config.get(n).connections.forall(c =>
                    config.get(n).Peers().exists(p => c.address == p)))

    //  if and address appears in connections and peers they should agree on the direction
    // TODO: might be broken
    val invDirectionsOut = modeledNodes.forall(n =>
        config.get(n).Peers().forall(p => 
        config.get(n).connections.forall(c => (c.address == p and c.state == "outgoing") 
                                        implies config.get(n).peerState.get(p) == "outbound")))

    // For nostalgic reasons and TLA+ lovers...
    val invTypeOK = modeledNodes.forall(n =>
        config.get(n).Peers().forall(p => 
        config.get(n).connections.forall(c =>
            connectionStates.contains(c.state) and
            peerStates.contains(config.get(n).peerState.get(p)))))

    // // for all modeled nodes y and all its peers x, some peerStates imply some connection state
    // val invConnectionsNotOpen = modeledNodes.forall(y => config.get(y).Peers().forall(x => 
    //     (config.get(y).peerState.get(x) == "dialing" or config.get(y).peerState.get(x) == "reconnecting") implies 
    //         config.get(y).connectionState.get(x) != "open"))

    // val invConnectionsNotInit = modeledNodes.forall(y => config.get(y).Peers().forall(x => 
    //     (config.get(y).peerState.get(x) == "inbound" or config.get(y).peerState.get(x) == "outbound") implies 
    //         config.get(y).connectionState.get(x) != "init"))

    // val invConnectionsOpen = modeledNodes.forall(y => config.get(y).activePeers().forall(x => 
    //     config.get(y).peerState.get(x) == "outbound" or config.get(y).peerState.get(x) == "inbound"))



/* ****************************************************************************
 * Actions
 * How the above functions can be invoked (e.g. by which node)
 * ************************************************************************* */

  
/* 
 * Future syntactic sugar.
 */
    // helps for better readability. 
    def require(pred:bool):bool = pred

/*
 * Useful actions to simplify spec
 */ 

    action allUnchanged = all {
        config' = config,
        Environment' = Environment,
        OperatorAddressBook' = OperatorAddressBook,
        requestingNodes' = requestingNodes,
    }    

    // Framerule
    action environmentUnchanged = all {
        Environment' = Environment,
        OperatorAddressBook' = OperatorAddressBook,
        requestingNodes' = requestingNodes,
    }    

    

/*
 * Local actions
 */

   // node dials a. This starts lower lever computations. 
    action actionDialPeer (node: Address, a:Address):bool = all {
        config' = config.set(node, config.get(node).dialPeer(a)),
        environmentUnchanged,
    }

    action actionTransportAccept (node: Address, a:Address):bool = all {
        config' = config.set(node, config.get(node).transportAccept(a)),
        environmentUnchanged,
    }

    action actionOnLostConnection(node: Address, c: Connection): bool = all {
        config' = config.set(node, config.get(node).onLostConnection(c)),
        environmentUnchanged,
    }

/* 
 * Environment actions: From the lower level (modeling the outcome of TCP connection
 * initialization, establishment, etc.)
 */

    // auxiliary functions
    pure def removeConnectionOutgoing(s: LocalState, a: Address): LocalState = {
        val y : Set[Connection] = s.connections.exclude(Set({address: a, state: "outgoing"}))
        s.with("connections", y)
    }

    pure def putPeer(s: LocalState, a: Address, direction: str): LocalState = {
        val y =  s.peerState.put(a, direction)
        s.with("peerState", y)
    }

    pure def loseConnection(s: LocalState, a: Address): LocalState = {
        val y : Set[Connection] = 
            s.connections.exclude(Set({address: a, state: "outgoing"}, {address: a, state: "incoming"}))
                         .union(Set({address: a, state: "lost"}))
        s.with("connections", y)
    }

 // return DialPeerSuccess -> if connection.address.id not in Peers -> peerState(address) = "outbound" 
 //                           otherwise -> delete connection from map 
 // in the spec right now: the peer might not exists, be already outbound (stutter), be reconnecting
 // TODO: check the cases
    action lowerLevelDialPeerSuccess(node: Address, a: Address): bool = all {
        require(config.get(node).connections.contains({address: a, state: "outgoing"})), 
        if (config.get(node).Peers().filter(x => x.ID == a.ID) == Set()) all {
            // If I don't have the ID in my peers already, add the address
            config' = config.set(node, config.get(node).putPeer(a, "outbound"))
        }
        else {
            // I have the ID already. That is, it might be incoming, or I might have dialed over a different
            // (network) address
            config' = config.set(node, config.get(node).removeConnectionOutgoing(a))
        },
        environmentUnchanged,
    }

    action lowerLevelDialPeerFailed(node: Address, a: Address): bool = all {
        require(config.get(node).connections.contains({address: a, state: "outgoing"})), 
        // TODO: perhaps add some checks
        config' = config.set(node, config.get(node).removeConnectionOutgoing(a)),
        environmentUnchanged,
    }

   action lowerLevelTransportAcceptSuccess(node: Address, a: Address): bool = all {
        require(config.get(node).connections.contains({address: a, state: "incoming"})), 
        if (config.get(node).Peers().filter(x => x.ID == a.ID) == Set()) all {
            // If I don't have the ID in my peers already, add the address
            config' = config.set(node, config.get(node).putPeer(a, "inbound"))
        }
        else {
            // I have the ID already. That is, it might be incoming, or I might have dialed over a different
            // (network) address
            config' = config.set(node, config.get(node).removeConnectionOutgoing(a))
        },
        environmentUnchanged,
    }

    action lowerLevelTransportAcceptFailed(node: Address, a: Address): bool = all {
        require(config.get(node).connections.contains({address: a, state: "incoming"})), 
        // TODO: perhaps add some checks
        config' = config.set(node, config.get(node).removeConnectionOutgoing(a)),
        environmentUnchanged,
    }

    action lowerLevelLoseConnection(node: Address, c: Connection):bool = all {
        require(config.get(node).connections.contains(c)),
        require(config.get(node).connections != Set()),
        // The following line is obscure magic and we discovered a bug in the cheat sheet
        config' = config.set(node, config.get(node).loseConnection(c.address)),
        environmentUnchanged,
    }

/* 
 * Upper level actions: callbacks from a reactor or from the operator
 */

    action reactorStopPeerForError(node: Address, peerID: str):bool = all {
        config' = config.set(node, config.get(node).stopPeerForError(peerID)),
        environmentUnchanged,
    }

    action operatorDialPeers(node:Address, addresses: Set[Address]):bool = all {
        config' = config.set(node, config.get(node).dialPeers(addresses)),
        environmentUnchanged,
    }

/* 
 * PEX flow
 */

    // pexRequest and lowerLevelPexAddresses abstract a flow where PEX realizes we need fresh Addresses
    // calls a seed node, and the seed node returns a set of addresses
    action pexRequest(node:Address): bool = all {
        // we do something only of addressbook is a subset of connections
        require(config.get(node).AddressBook.filter(a => not(config.get(node).connections.exists(c => c.address == a))) == Set()),
        requestingNodes' = requestingNodes.union(Set(node)),
        config' = config,
        Environment' = Environment,
        OperatorAddressBook' = OperatorAddressBook,
    }

    // not sure if the caller of this is "lower lever" or the PEX reactor
    action lowerLevelPexAddresses(node:Address, addresses: Set[Address]):bool = all {
        require(requestingNodes.contains(node)),
        config' = config.set(node, config.get(node).PexAddrs(addresses)),
        requestingNodes' = requestingNodes.exclude(Set(node)),
        Environment' = Environment,
        OperatorAddressBook' = OperatorAddressBook,
    }


    pure def ensurePeers(s: LocalState, addr: Address): LocalState = {
        if (s.peersOutbound() + s.peersDialing() < s.MaxNumOutboundPeers) {
            s.dialPeer(addr)
        }
        else {
            s
        }
    }

    action pexEnsurePeers(node: Address): bool = all {
        //val addr = s.AddressBook.chooseSome() // addressbook.pickAddress(Bias)
        nondet addr = oneOf(config.get(node).AddressBook)
        config' = config.set(node, config.get(node).ensurePeers(addr)),
        environmentUnchanged,
    }



    
/* ****************************************************************************
 * Initial states (potentially many for different scenarios)
 * Steps (potentially many modelling different environments, e.g. including
          faulty peers)
 * Tests (specific runs with assertions)
 * ************************************************************************* */



    /*
    Here we start to describe initial states and runs and tests
    */

    action init = all {
        OperatorAddressBook' = Set({ID: "alice", networkAddress: "127.0.0.111:10"}, {ID: "bob", networkAddress: "127.0.0.112:10"}),
        Environment' = Set({ID: "lasaro", networkAddress: "127.0.0.7:10"}, {ID: "daniel", networkAddress: "127.0.0.6:10"}),
        requestingNodes' = Set(),
        config' = Map (
            {ID: "scully", networkAddress: "127.0.0.1177:10"} -> {
                peerState: Map(),
                connections: Set(),
                AddressBook: Set({ID: "igor", networkAddress: "127.0.0.1:10"}, {ID: "josef", networkAddress: "127.0.0.2:10"}),
                IsUnconditional: Set({ID: "zarko", networkAddress: "127.0.0.42:10"}),
                PersistentPeers: Set({ID: "bucky", networkAddress: "127.0.0.9:10"}),
                MaxNumOutboundPeers: 5,
                MaxNumInboundPeers: 10,
            }
        )
    }

    action initEnvEquAddrBook = all {
        val add = Set({ID: "lasaro", networkAddress: "127.0.0.7:10"}, {ID: "daniel", networkAddress: "127.0.0.6:10"})
        all {
        OperatorAddressBook' = Set({ID: "alice", networkAddress: "127.0.0.111:10"}, {ID: "bob", networkAddress: "127.0.0.112:10"}),
        Environment' = add,
        requestingNodes' = Set(),
        config' = Map (
            {ID: "scully", networkAddress: "127.0.0.1177:10"} -> {
                peerState: Map(),
                connections: Set(),
                AddressBook: add,
                IsUnconditional: Set({ID: "zarko", networkAddress: "127.0.0.42:10"}),
                PersistentPeers: Set({ID: "bucky", networkAddress: "127.0.0.9:10"}),
                MaxNumOutboundPeers: 5,
                MaxNumInboundPeers: 10,
            },
             {ID: "mulder", networkAddress: "127.0.0.1177:210"} -> {
                peerState: Map(),
                connections: Set(),
                AddressBook: add,
                IsUnconditional: Set({ID: "alien", networkAddress: "127.0.0.142:10"}),
                PersistentPeers: Set({ID: "monster", networkAddress: "127.0.0.19:10"}),
                MaxNumOutboundPeers: 5,
                MaxNumInboundPeers: 10,
            }
        )
        }
    }


    run happyPathTest = {
        val a = {ID: "alice", networkAddress: "127.0.0.111:10"}
        val node = {ID: "scully", networkAddress: "127.0.0.1177:10"}
        init
        .then(operatorDialPeers(node , Set(a)))
        .then(lowerLevelDialPeerSuccess(node, a))
        .then(all{
            assert(
                and {
                    config.get(node).connections.contains({address: a, state: "outgoing"}),
                    config.get(node).peerState.get(a) == "outbound",
                }
            ),
            allUnchanged,
        })
    }

    run mixedDirectionsFailsTest = {
        val a = {ID: "alice", networkAddress: "127.0.0.111:10"}
        val node = {ID: "scully", networkAddress: "127.0.0.1177:10"}
        init
        .then(actionDialPeer(node,a))
        .then(actionTransportAccept(node,a))
        .then(lowerLevelTransportAcceptSuccess(node,a))
        .then(all{
            invDirectionsOut,
            allUnchanged,
        })
        .fail()
    }

    action step = any {
        nondet a = oneOf(modeledNodes) 
        nondet outgoing = oneOf(config.get(a).AddressBook)
        nondet incoming = oneOf(Environment)
        nondet conn = oneOf(config.get(a).connections.union(Set({address: {ID: "default", networkAddress: "0.0.0.0"}, 
                                                                 state: "outgoing"})))
        nondet dial = oneOf(OperatorAddressBook.powerset())
        nondet seedResponse = oneOf(Environment.powerset())
            any {            
                actionDialPeer(a, outgoing), // create a rawConnection
                lowerLevelDialPeerFailed(a, outgoing),
                lowerLevelDialPeerSuccess(a, outgoing), // if dialed node modeled, we require Transportaccept

                // action to add rawConnection where first element is not in modeledNodes

                actionTransportAccept(a, incoming), // guarded by existence of rawconnection where a is second element 
                lowerLevelTransportAcceptFailed(a, incoming),
                lowerLevelTransportAcceptSuccess(a, incoming),

                // TODO: conn empty
                lowerLevelLoseConnection(a, conn),
                actionOnLostConnection(a, conn),                

                // TODO: pick peerID from Peers()
                reactorStopPeerForError(a, "test"),

                operatorDialPeers(a, dial),

                pexEnsurePeers(a),

                pexRequest(a),
                lowerLevelPexAddresses(a, seedResponse)
            },
    }
   

    





}
