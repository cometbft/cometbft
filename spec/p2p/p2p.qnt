module p2p {

    pure val peerStates = Set("dialing", "inbound", "reconnecting", "outbound", "unconditional")
    // a work around as constants are not working in the REPL

    pure val connectionStates = Set("init", "open", "closed")

    type Address = {ID:str, ipAddress:str}
 //   type State = {peerState:str, connectionState: str}

    //type Peer = {address:Address, state:State}
    //var Peers : Set[Peer]

    // p2p variables 
    var peerState: Address -> str
    var connectionState: Address -> str
    var IsUnconditional: Set[Address]

    // PEX variable
    var AddressBook: Set[Address]

    // Environment. These are the nodes that may dial us
    var Environment: Set[Address]
 

    pure val MaxNumOutboundPeers = 5
    pure val MaxNumInboundPeers = 10

    def Peers = peerState.keys()

    def peersOutbound = size(Peers.filter(a => peerState.get(a) == "outbound"))
    def peersInbound = size(Peers.filter(a => peerState.get(a) == "inbound"))
    def peersDialing = size(Peers.filter(a => peerState.get(a) == "dialing"))

    def require(pred:bool):bool = pred
    //pure def putset (m,s)

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
        peerState' = peerState.put(a, "inbound"),  //shall we check if it is already outbound and open?
        connectionState' = connectionState.put(a, "open"),
        AddressBook' = AddressBook.union(Set(a)) ,
        all { //unchanged
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
    }
        // inform all reactors then
    

    action ensurePeers(a:Address):bool =
        all {
                require(peersOutbound + peersDialing < MaxNumOutboundPeers),
                require(not(Peers.contains(a))),
                all {
                    // I guess we don't remove AddressBook' = AddressBook.exclude(Set(a)),
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

    action dialPeers(S:Set[Address]):bool = // check what the code actually is doing. Now we don't override existing peers in this spec
        all {
            peerState' = Peers.union(S).mapBy(a => if (Peers.contains(a)) peerState.get(a) else "dialing"), 
            connectionState' = Peers.union(S).mapBy(a => if (Peers.contains(a)) connectionState.get(a) else "init"),
            all { //unchanged
                        AddressBook' = AddressBook,
                        Environment' = Environment,
                        IsUnconditional' = IsUnconditional,
                    }
        }

    action init =
        all {
            peerState' = Map(),
            connectionState' = Map(),
            AddressBook' = Set({ID: "igor", ipAddress: "127.0.0.1"}, {ID: "josef", ipAddress: "127.0.0.2"}),
            Environment' = Set({ID: "lasaro", ipAddress: "127.0.0.7"}, {ID: "daniel", ipAddress: "127.0.0.6"}),
            IsUnconditional' = Set({ID: "zarko", ipAddress: "127.0.0.42"}),
        }

    action step = any {
        all {
            require(AddressBook.exclude(Peers) != Set()),
            nondet a = oneOf(AddressBook.exclude(Peers)) // addrbook.PickAddress(bias)
            ensurePeers(a),
        },
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