//quint test gossip_ref.qnt --main P2PTests --verbosity=5
module P2PTests {
    import Helpers.* from "./Helpers"
    import Globals.* from "./Globals"

    //XOR import
        import P2PInstCheat as P2PInst from "./P2PInstCheat"
        //import P2PInstNetwork as P2PInst from "./P2PInstNetwork"

    action unchangedAll = all {
        P2PInst::PA::unchangedAll,
    }

    action init = all {
        P2PInst::PA::init,
    }

    //===== Runs/Debug
    run sendReceiveInRingTopTest: bool = {
        init
            //connect some nodes
            .then(P2PInst::PA::connectRingStep)
            //send and receive messages
            .then(
                all {
                    P2PInst::send(1,2,{tag: "query_request"}),

                    P2PInst::PA::pid2peerState' = P2PInst::PA::pid2peerState,
                }
            )
            .then(
                all {
                    P2PInst::send(2,3,{tag: "query_request"}),

                    P2PInst::PA::pid2peerState' = P2PInst::PA::pid2peerState,
                }
            )
            .then(P2PInst::receiveMessageStep.repeated(2))
            .then(all {
                //RGPUnicastProp,
                P2PInst::PA::RGPConcurrent,
                P2PInst::PA::RGPIgnoring,
                unchangedAll
            })
            .then(
                all {
                    P2PInst::send(1,3,{tag: "query_request"}),

                    P2PInst::PA::pid2peerState' = P2PInst::PA::pid2peerState,
                }.fail())
    }

    run sendReceiveInFullyConnectedTest: bool = {
        init
            //connect some nodes
            .then(P2PInst::PA::connectFullStep)
            //send and receive messages
            .then(
                all {
                    P2PInst::send(1,2,{tag: "query_request"}),

                    P2PInst::PA::pid2peerState' = P2PInst::PA::pid2peerState,
                }
            )
            .then(
                all {
                    P2PInst::send(2,3,{tag: "query_request"}),

                    P2PInst::PA::pid2peerState' = P2PInst::PA::pid2peerState,
                }
            )
            .then(
                all {
                    P2PInst::send(1,3,{tag: "query_request"}),

                    P2PInst::PA::pid2peerState' = P2PInst::PA::pid2peerState,
                }
            )
            .then(P2PInst::receiveMessageStep.repeated(3))
            .then(all {
                //RGPUnicastProp,
                P2PInst::PA::RGPConcurrent,
                P2PInst::PA::RGPIgnoring,
                unchangedAll
            })
    }

    run sendReceiveInStarConnectedTest: bool = {
        init
            //connect some nodes
            .then(P2PInst::PA::connectStarStep)
            //send and receive messages
            .then(
                all {
                    P2PInst::send(1,2,{tag: "query_request"}),

                    P2PInst::PA::pid2peerState' = P2PInst::PA::pid2peerState,
                }
            )
            .then(
                all {
                    P2PInst::send(1,3,{tag: "query_request"}),

                    P2PInst::PA::pid2peerState' = P2PInst::PA::pid2peerState,
                }
            )
            .then(
                all {
                    P2PInst::send(1,5,{tag: "query_request"}),

                    P2PInst::PA::pid2peerState' = P2PInst::PA::pid2peerState,
                }
            )
            .then(P2PInst::receiveMessageStep.repeated(3))
            .then(all {
                //RGPUnicastProp,
                P2PInst::PA::RGPConcurrent,
                P2PInst::PA::RGPIgnoring,
                unchangedAll
            })
            .then(
                all {
                    P2PInst::send(2,3,{tag: "query_request"}),

                    P2PInst::PA::pid2peerState' = P2PInst::PA::pid2peerState,
                }.fail()
            )
    }

    run sendReceiveInRandomConnectedTest: bool = {
        init
            //connect some nodes
            .then(P2PInst::PA::connectRandomStep)
            //send and receive messages
            .then(
                all {
                    nondet p = P2PInst::PA::pid2peerState.keys().filter(pa => P2PInst::PA::pid2peerState.get(pa).connectedNodes.size() > 0).oneOf()
                    nondet q = P2PInst::PA::pid2peerState.get(p).connectedNodes.oneOf()
                    P2PInst::send(p,q,{tag: "query_request"}),
                    P2PInst::PA::pid2peerState' = P2PInst::PA::pid2peerState,
                }.repeated(1)
            )
            .then(P2PInst::receiveMessageStep.repeated(1))
            .then(all {
                //RGPUnicastProp,
                P2PInst::PA::RGPConcurrent,
                P2PInst::PA::RGPIgnoring,
                unchangedAll
            })
    }
}