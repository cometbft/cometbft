//==== Gossip Tests
module GossipTests {
    import Helpers.* from "./Helpers"
    import Globals.* from "./Globals"

    //XOR import
        //import GossipInstCheat as GInst from "./GossipInstCheat"
        //import GossipInstNoP2P as GInst from "./GossipInstNoP2P"
        import GossipInstP2P as GInst from "./GossipInstP2P"

    var stepCounter: int
    val stateAll = {
        stepCounter:stepCounter,
        pa:GInst::GA::P2PInst::PA::stateAll,
        pa:GInst::GA::stateAll
    }

    action init = all {
        stepCounter' = 1
    }

    def gossipablePairs =
        tuples(ProcIds,ProcIds).filter(t => and (
            t._1 != t._2,
            GInst::GA::localView(t._1) != GInst::GA::localView(t._2),
            GInst::GA::areConnected(t._1, t._2)
    ))

//===== Runs/Debug
    run gossipInRingTopoTest: bool = {
        val signer = 1
        val entry = {tag:"proposal", height:1, round:0, payload:"PROPOSAL", signer: signer}

        (all {
            init,
            GInst::init,
            GInst::GA::init,
            GInst::GA::P2PInst::PA::init,
        })
        //connect some nodes
        .then(
                all {
                    GInst::GA::P2PInst::PA::connectRingStep,
                    GInst::GA::unchangedAll,
                    GInst::unchangedAll,
                    stepCounter' = stepCounter
        })
        //add one entry
        .then(all {
            GInst::addTuple(signer, entry),
            GInst::GA::P2PInst::PA::unchangedAll,
            GInst::unchangedAll,
            stepCounter' = stepCounter
        })
        //gossip
        .then((all {
            print("outer step", stepCounter),

            val pqs = gossipablePairs
            if (pqs.size() > 0) all {
                stepCounter' = stepCounter + 1,
                print ("step yeah", {pairs:pqs, stateAll:stateAll}),

                any {
                    all {
                        //Send or process received messages.
                        nondet pq = pqs.oneOf()
                        GInst::gossipStep(pq._1, pq._2),
                    },
                    all {
                        //Deliver messages
                        GInst::GA::unchangedAll,
                        GInst::GA::P2PInst::receiveMessageStep,
                        GInst::unchangedAll,
                    }
                },
            } else all {
                print ("step nope", {pairs:pqs, stateAll:stateAll}),

                //Either it progresses or all have the entry.
                ProcIds.forall(p => GInst::GA::localView(p).contains(entry)),

                GInst::GA::P2PInst::PA::unchangedAll,
                GInst::GA::unchangedAll,
                GInst::unchangedAll,
                stepCounter' = stepCounter,
            },
        }).repeated(ProcIds.size()*2*7))
        .then( all {
            //All see the entry
            ProcIds.forall(p => GInst::GA::localView(p).contains(entry)),

            GInst::GA::P2PInst::PA::unchangedAll,
            GInst::GA::unchangedAll,
            GInst::unchangedAll,
            stepCounter' = stepCounter
        })
    }

    run gossipCliqueTopoTest: bool = {
        val signer = 1
        val entry = {tag:"proposal", height:1, round:0, payload:"PROPOSAL", signer: signer}

        (all {
            init,
            GInst::init,
            GInst::GA::init,
            GInst::GA::P2PInst::PA::init,
        })
        //connect some nodes
        .then(
        all {
            GInst::GA::P2PInst::PA::connectFullStep,
            GInst::GA::unchangedAll,
            GInst::unchangedAll,
            stepCounter' = stepCounter
        })
        //add one entry
        .then(all {
            GInst::addTuple(signer, entry),
            GInst::GA::P2PInst::PA::unchangedAll,
            GInst::unchangedAll,
            stepCounter' = stepCounter
        })
        //gossip
        .then((all {
            print("outer step", stepCounter),

            val pqs = gossipablePairs
            if (pqs.size() > 0) all {
                stepCounter' = stepCounter + 1,
                print ("step yeah", {pairs:pqs, stateAll:stateAll}),

                any {
                    all {
                        //Send or process received messages.
                        nondet pq = pqs.oneOf()
                        GInst::gossipStep(pq._1, pq._2),

                    },
                    all {
                        //Deliver messages
                        GInst::GA::unchangedAll,
                        GInst::GA::P2PInst::receiveMessageStep,
                        GInst::unchangedAll,
                    }
                },
            } else all {
                print ("step nope", {pairs:pqs, stateAll:stateAll}),

                //Either it progresses or all have the entry.
                ProcIds.forall(p => GInst::GA::localView(p).contains(entry)),

                GInst::GA::P2PInst::PA::unchangedAll,
                GInst::GA::unchangedAll,
                GInst::unchangedAll,
                stepCounter' = stepCounter,

            },
        }).repeated(ProcIds.size()*2*7))
        .then( all {
            //All see the entry
            ProcIds.forall(p => GInst::GA::localView(p).contains(entry)),

            GInst::GA::P2PInst::PA::unchangedAll,
            GInst::GA::unchangedAll,
            GInst::unchangedAll,
            stepCounter' = stepCounter,

        })
    }

    run gossipInStarTopoTest: bool = {
            val signer = 1
        val entry = {tag:"proposal", height:1, round:0, payload:"PROPOSAL", signer: signer}

        (all {
            init,
            GInst::init,
            GInst::GA::init,
            GInst::GA::P2PInst::PA::init,
        })
        //connect some nodes
        .then(
                all {
                    GInst::GA::P2PInst::PA::connectStarStep,
                    GInst::GA::unchangedAll,
                    GInst::unchangedAll,
                    stepCounter' = stepCounter
        })
        //add one entry
        .then(all {
            GInst::addTuple(signer, entry),
            GInst::GA::P2PInst::PA::unchangedAll,
            GInst::unchangedAll,
            stepCounter' = stepCounter
        })
        //gossip
        .then((all {
            print("outer step", stepCounter),

            val pqs = gossipablePairs
            if (pqs.size() > 0) all {
                stepCounter' = stepCounter + 1,
                print ("step yeah", {pairs:pqs, stateAll:stateAll}),

                any {
                    all {
                        //Send or process received messages.
                        nondet pq = pqs.oneOf()
                        GInst::gossipStep(pq._1, pq._2),
                    },
                    all {
                        //Deliver messages
                        GInst::GA::unchangedAll,
                        GInst::GA::P2PInst::receiveMessageStep,
                        GInst::unchangedAll,
                    }
                },
            } else all {
                print ("step nope", {pairs:pqs, stateAll:stateAll}),

                //Either it progresses or all have the entry.
                ProcIds.forall(p => GInst::GA::localView(p).contains(entry)),

                GInst::GA::P2PInst::PA::unchangedAll,
                GInst::GA::unchangedAll,
                GInst::unchangedAll,
                stepCounter' = stepCounter,

            },
        }).repeated(ProcIds.size()*2*7))
        .then( all {
            //All see the entry
            ProcIds.forall(p => GInst::GA::localView(p).contains(entry)),

            GInst::GA::P2PInst::PA::unchangedAll,
            GInst::GA::unchangedAll,
            GInst::unchangedAll,
            stepCounter' = stepCounter,

        })
    }

    //TODO: reenable test once the topoloy is build correctly.
    action gossipInRandomTopoTest: bool = {
            val signer = 1
        val entry = {tag:"proposal", height:1, round:0, payload:"PROPOSAL", signer: signer}

        (all {
            init,
            GInst::init,
            GInst::GA::init,
            GInst::GA::P2PInst::PA::init,
        })
        //connect some nodes
        .then(
                all {
                    GInst::GA::P2PInst::PA::connectRandomStep,
                    GInst::GA::unchangedAll,
                    GInst::unchangedAll,
                    stepCounter' = stepCounter
        })
        //add one entry
        .then(all {
            GInst::addTuple(signer, entry),
            GInst::GA::P2PInst::PA::unchangedAll,
            GInst::unchangedAll,
            stepCounter' = stepCounter
        })
        //gossip
        .then((all {
            print("outer step", stepCounter),

            val pqs = gossipablePairs
            if (pqs.size() > 0) all {
                stepCounter' = stepCounter + 1,
                print ("step yeah", {pairs:pqs, stateAll:stateAll}),

                any {
                    all {
                        //Send or process received messages.
                        nondet pq = pqs.oneOf()
                        GInst::gossipStep(pq._1, pq._2),
                    },
                    all {
                        //Deliver messages
                        GInst::GA::unchangedAll,
                        GInst::GA::P2PInst::receiveMessageStep,
                        GInst::unchangedAll,
                    }
                },
            } else all {
                print ("step nope", {pairs:pqs, stateAll:stateAll}),

                //Either it progresses or all have the entry.
                ProcIds.forall(p => GInst::GA::localView(p).contains(entry)),

                GInst::GA::P2PInst::PA::unchangedAll,
                GInst::GA::unchangedAll,
                GInst::unchangedAll,
                stepCounter' = stepCounter,

            },
        }).repeated(ProcIds.size()*2*7))
        .then( all {
            //All see the entry
            ProcIds.forall(p => GInst::GA::localView(p).contains(entry)),

            GInst::GA::P2PInst::PA::unchangedAll,
            GInst::GA::unchangedAll,
            GInst::unchangedAll,
            stepCounter' = stepCounter,
        })
    }
}