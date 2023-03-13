module p2p {

    // TODO: "pure val" is a work around as constants are not working in the REPL, yet
    pure val peerStates = Set("dialing", "inbound", "reconnecting", "outbound", "unconditional")
    // TODO: Do we actually care about connection states at this level of abstraction?
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

    // Environment. These are the nodes that may dial us
    var Environment: Set[Address]
 
    def Peers = peerState.keys()

    // counting peers in different states
    def peersOutbound = size(Peers.filter(a => peerState.get(a) == "outbound"))
    def peersInbound = size(Peers.filter(a => peerState.get(a) == "inbound"))
    def peersDialing = size(Peers.filter(a => peerState.get(a) == "dialing"))

    // helps for better readability. Future syntactic sugar 
    def require(pred:bool):bool = pred
    // TODO: pure def putset (m,s)

    // triggered by "the network"
    action loseConnection(a:Address):bool = all{
        require(connectionState.get(a) != "closed"),
        connectionState' = connectionState.set(a, "closed"),
        all { //unchanged
            AddressBook' = AddressBook,
            peerState' = peerState,
            Environment' = Environment,
            IsUnconditional' = IsUnconditional,
        }
    }

    action transportAccept(a: Address):bool = all {
        or {
            peersInbound < MaxNumInboundPeers,
            IsUnconditional.contains(a)
        },
        require(peerState.get(a) != "outbound"), // TODO: check whether is is the right condition
        peerState' = peerState.put(a, "inbound"), 
        connectionState' = connectionState.put(a, "open"),
        AddressBook' = AddressBook.union(Set(a)) ,
        all { //unchanged
            Environment' = Environment,
            IsUnconditional' = IsUnconditional,
        }
    }

    // lower level connection initialization complete. Gloss over details
    // TODO: what happens if the other side doesn't accept the connection of some problem occurs?
    action establishConnection(a:Address):bool = all{ // TODO: upgradeConnection?
        require(connectionState.get(a) != "dialing"),
        connectionState' = connectionState.set(a, "outbound"),
        all { //unchanged
            AddressBook' = AddressBook,
            peerState' = peerState,
            Environment' = Environment,
            IsUnconditional' = IsUnconditional,
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
        }
        // TODO: inform all reactors then. (How) should we model that?
    }

    action debugEnsurePeers(a:Address):bool =
        all {
                require(peersOutbound + peersDialing < MaxNumOutboundPeers),
                require(not(Peers.contains(a))),
                all {
                    //dialPeer(addr)
                    peerState' = peerState.put(a, "dialing"), 
                    connectionState' = connectionState.put(a, "init"),
                    all { //unchanged
                        AddressBook' = AddressBook,
                        Environment' = Environment,
                        IsUnconditional' = IsUnconditional,
                    }
                }
            }

        
    action ensurePeers = all {
            require(AddressBook.exclude(Peers) != Set()),
            nondet a = oneOf(AddressBook.exclude(Peers)) // addrbook.PickAddress(bias)
            debugEnsurePeers(a),
        }


    action dialPeers(S:Set[Address]):bool = // check what the code actually is doing. Now we don't override existing peers in this spec
        all {
            peerState' = Peers.union(S).mapBy(a => if (Peers.contains(a)) peerState.get(a) else "dialing"), 
            connectionState' = Peers.union(S).mapBy(a => if (Peers.contains(a)) connectionState.get(a) else "init"),
            AddressBook' = AddressBook.union(S),
            all { //unchanged
                        
                        Environment' = Environment,
                        IsUnconditional' = IsUnconditional,
                    }
        }

    action init =
        all {
            peerState' = Map(),
            connectionState' = Map(),
            AddressBook' = Set({ID: "igor", networkAddress: "127.0.0.1:10"}, {ID: "josef", networkAddress: "127.0.0.2:10"}),
            Environment' = Set({ID: "lasaro", networkAddress: "127.0.0.7:10"}, {ID: "daniel", networkAddress: "127.0.0.6:10"}),
            IsUnconditional' = Set({ID: "zarko", networkAddress: "127.0.0.42"}),
        }

    action initWithSameSets =
        val initSet = Set({ID: "igor", networkAddress: "127.0.0.1:10"}, {ID: "josef", networkAddress: "127.0.0.2:10"})
        all {
            peerState' = Map(),
            connectionState' = Map(),
            AddressBook' = initSet,
            Environment' = initSet,
            IsUnconditional' = Set({ID: "zarko", networkAddress: "127.0.0.42"}),
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
            nondet a = oneOf(Environment) 
            transportAccept(a),
        }

    }

    val invPeers = peerState.keys() == connectionState.keys()


}